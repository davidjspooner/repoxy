package container

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type containerUpstreamAuth struct {
	bearer *bearerTokenSource
	basic  basicCredentialSource
}

func newContainerUpstreamAuth(httpClient client.Interface, upstream repo.Upstream) (*containerUpstreamAuth, error) {
	auth := &containerUpstreamAuth{
		bearer: newBearerTokenSource(httpClient, "", ""),
	}
	if upstream.Auth == nil {
		return auth, nil
	}
	provider := strings.ToLower(upstream.Auth.Provider)
	cfg := upstream.Auth.Config
	switch provider {
	case "", "dockerhub", "ghcr", "bearer":
		username := ""
		password := ""
		if cfg != nil {
			username = cfg["username"]
			password = cfg["password"]
		}
		auth.bearer = newBearerTokenSource(httpClient, username, password)
	case "basic":
		if cfg == nil {
			return nil, fmt.Errorf("basic upstream auth requires username and password")
		}
		user := cfg["username"]
		pass := cfg["password"]
		if user == "" || pass == "" {
			return nil, fmt.Errorf("basic upstream auth requires username and password")
		}
		auth.basic = &staticBasicCredentials{
			value: base64.StdEncoding.EncodeToString([]byte(user + ":" + pass)),
		}
	case "ecr":
		creds, err := newECRBasicCredentials(cfg)
		if err != nil {
			return nil, err
		}
		auth.basic = creds
	default:
		return nil, fmt.Errorf("unsupported upstream auth provider %q", upstream.Auth.Provider)
	}
	return auth, nil
}

func (a *containerUpstreamAuth) authorization(resp *http.Response) (string, error) {
	if resp == nil {
		return "", nil
	}
	challengeHeader := resp.Header.Get("WWW-Authenticate")
	challengeHeader = strings.TrimSpace(challengeHeader)
	if challengeHeader == "" {
		return "", nil
	}
	challenges, err := client.ParseWWWAuthenticate(resp.Request.Context(), challengeHeader)
	if err != nil {
		if challenge := parseBearerChallengeFallback(challengeHeader); challenge != nil {
			challenges = []client.Challenge{*challenge}
		} else {
			return "", err
		}
	}
	for _, challenge := range challenges {
		switch strings.ToLower(challenge.Scheme) {
		case "bearer":
			if a.bearer == nil {
				continue
			}
			token, err := a.bearer.Token(resp.Request.Context(), challenge)
			if err != nil {
				return "", err
			}
			if token != "" {
				return "Bearer " + token, nil
			}
		case "basic":
			if a.basic == nil {
				continue
			}
			headerValue, err := a.basic.HeaderValue(resp.Request.Context())
			if err != nil {
				return "", err
			}
			if headerValue != "" {
				return "Basic " + headerValue, nil
			}
		}
	}
	return "", nil
}

func parseBearerChallengeFallback(header string) *client.Challenge {
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return nil
	}
	paramStr := strings.TrimSpace(header[len("bearer"):])
	paramStr = strings.TrimSpace(paramStr)
	if paramStr == "" {
		return &client.Challenge{Scheme: "Bearer", Params: map[string]string{}}
	}
	params := map[string]string{}
	for _, part := range strings.Split(paramStr, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		if key != "" {
			params[strings.ToLower(key)] = val
		}
	}
	return &client.Challenge{Scheme: "Bearer", Params: params}
}

type bearerTokenSource struct {
	client   client.Interface
	username string
	password string

	mu    sync.Mutex
	cache map[string]*bearerToken
}

type bearerToken struct {
	value     string
	expiresAt time.Time
}

func newBearerTokenSource(httpClient client.Interface, username, password string) *bearerTokenSource {
	return &bearerTokenSource{
		client:   httpClient,
		username: username,
		password: password,
		cache:    map[string]*bearerToken{},
	}
}

func (s *bearerTokenSource) Token(ctx context.Context, challenge client.Challenge) (string, error) {
	if challenge.Params == nil {
		return "", fmt.Errorf("missing bearer params")
	}
	realm := challenge.Params["realm"]
	if realm == "" {
		return "", fmt.Errorf("missing realm in challenge")
	}
	service := challenge.Params["service"]
	scope := challenge.Params["scope"]
	cacheKey := service + "|" + scope
	if token := s.cached(cacheKey); token != "" {
		return token, nil
	}
	u, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("invalid realm %q: %w", realm, err)
	}
	q := u.Query()
	if service != "" {
		q.Set("service", service)
	}
	if scope != "" {
		q.Set("scope", scope)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	if s.username != "" {
		req.SetBasicAuth(s.username, s.password)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("token endpoint returned %s: %s", resp.Status, string(body))
	}
	var payload struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode bearer token response: %w", err)
	}
	token := payload.Token
	if token == "" {
		token = payload.AccessToken
	}
	if token == "" {
		return "", fmt.Errorf("token response missing token field")
	}
	expiry := time.Now().Add(5 * time.Minute)
	if payload.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	s.store(cacheKey, token, expiry)
	return token, nil
}

func (s *bearerTokenSource) cached(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.cache[key]; ok {
		if time.Now().Before(entry.expiresAt.Add(-10 * time.Second)) {
			return entry.value
		}
		delete(s.cache, key)
	}
	return ""
}

func (s *bearerTokenSource) store(key, token string, expires time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = &bearerToken{value: token, expiresAt: expires}
}

type basicCredentialSource interface {
	HeaderValue(ctx context.Context) (string, error)
}

type staticBasicCredentials struct {
	value string
}

func (s *staticBasicCredentials) HeaderValue(context.Context) (string, error) {
	return s.value, nil
}

type ecrAPI interface {
	GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error)
}

var newECRClient = func(cfg aws.Config) ecrAPI {
	return awsecr.NewFromConfig(cfg)
}

type ecrBasicCredentials struct {
	client     ecrAPI
	registryID string
	now        func() time.Time

	mu      sync.Mutex
	cached  string
	expires time.Time
}

func newECRBasicCredentials(cfg map[string]string) (*ecrBasicCredentials, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ecr auth requires configuration")
	}
	region := cfg["region"]
	accessKey := cfg["access_key_id"]
	secretKey := cfg["secret_access_key"]
	if region == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("ecr auth requires region, access_key_id, and secret_access_key")
	}
	session := cfg["session_token"]
	awsCfg := aws.Config{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, session)),
	}
	return &ecrBasicCredentials{
		client:     newECRClient(awsCfg),
		registryID: cfg["registry_id"],
		now:        time.Now,
	}, nil
}

func (e *ecrBasicCredentials) HeaderValue(ctx context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cached != "" && e.now().Before(e.expires.Add(-1*time.Minute)) {
		return e.cached, nil
	}
	input := &awsecr.GetAuthorizationTokenInput{}
	if e.registryID != "" {
		input.RegistryIds = append(input.RegistryIds, e.registryID)
	}
	resp, err := e.client.GetAuthorizationToken(ctx, input)
	if err != nil {
		return "", fmt.Errorf("fetch ecr authorization token: %w", err)
	}
	if len(resp.AuthorizationData) == 0 {
		return "", fmt.Errorf("ecr authorization response missing data")
	}
	data := resp.AuthorizationData[0]
	token := aws.ToString(data.AuthorizationToken)
	if token == "" {
		return "", fmt.Errorf("ecr authorization token empty")
	}
	expiration := aws.ToTime(data.ExpiresAt)
	if expiration.IsZero() {
		expiration = e.now().Add(12 * time.Hour)
	}
	e.cached = token
	e.expires = expiration
	return token, nil
}
