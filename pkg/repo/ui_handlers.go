package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/davidjspooner/go-http-server/pkg/mux"
)

// RegisterUIRoutes attaches the read-only UI API endpoints to the mux.
func RegisterUIRoutes(m *mux.ServeMux) {
	m.HandleFunc("GET /api/ui/v1/repository-types", handleListTypes)
	m.HandleFunc("GET /api/ui/v1/repository-types/{typeId}/repositories", handleListRepositoriesForType)
	m.HandleFunc("GET /api/ui/v1/repositories/{repoId}/items", handleListItems)
	m.HandleFunc("GET /api/ui/v1/items/{itemId...}/versions", handleListVersions)
	m.HandleFunc("GET /api/ui/v1/versions/{versionId...}/files", handleListFiles)
	m.HandleFunc("GET /api/ui/v1/files/{fileId...}", handleFileDetail)
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	var body errorBody
	body.Error.Code = code
	body.Error.Message = msg
	writeJSON(w, status, body)
}

func handleListTypes(w http.ResponseWriter, r *http.Request) {
	rTypeLock.RLock()
	defer rTypeLock.RUnlock()

	typeBody := struct {
		Types []TypeMeta `json:"types"`
	}{}
	for _, td := range rTypeDetails {
		meta := td.rType.Meta()
		if meta.ID == "" {
			meta.ID = td.typeName
		}
		if td.typeName == "tofu" || strings.EqualFold(meta.ID, "tofu") {
			if meta.Label == "" || strings.EqualFold(meta.Label, "terraform") {
				meta.Label = "OpenTofu"
			}
			if meta.Description == "" {
				meta.Description = "Providers delivered from pull-through caches of registry.opentofu.org or private catalogs"
			}
		} else if meta.Label == "" {
			meta.Label = cases.Title(language.English).String(td.typeName)
		}
		typeBody.Types = append(typeBody.Types, meta)
	}
	sort.Slice(typeBody.Types, func(i, j int) bool { return typeBody.Types[i].ID < typeBody.Types[j].ID })
	writeJSON(w, http.StatusOK, typeBody)
}

func handleListRepositoriesForType(w http.ResponseWriter, r *http.Request) {
	typeID := r.PathValue("typeId")
	td := getTypeByID(typeID)
	if td == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository type not found")
		return
	}
	typeMeta := td.rType.Meta()
	if typeMeta.ID == "" {
		typeMeta.ID = td.typeName
	}
	var repos []InstanceMeta
	for _, inst := range td.instances {
		if inst == nil || inst.instance == nil {
			continue
		}
		repoMeta := inst.instance.Describe()
		if repoMeta.TypeID == "" {
			repoMeta.TypeID = typeMeta.ID
		}
		repos = append(repos, repoMeta)
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].ID < repos[j].ID })

	writeJSON(w, http.StatusOK, map[string]any{
		"type": map[string]any{
			"id":    typeMeta.ID,
			"label": typeMeta.Label,
		},
		"repositories": repos,
	})
}

func handleListItems(w http.ResponseWriter, r *http.Request) {
	repoID := r.PathValue("repoId")
	inst := getInstanceByRepoID(repoID)
	if inst == nil || inst.common == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found")
		return
	}
	store := inst.common
	ctx := r.Context()
	hosts, err := store.ListHosts(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	var items []map[string]any
	for _, host := range hosts {
		names, err := store.ListNamesForHost(ctx, host)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
			return
		}
		for _, name := range names {
			id := repoID + ":" + host + "/" + name
			items = append(items, map[string]any{
				"id":     id,
				"label":  name,
				"detail": host,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i]["id"].(string) < items[j]["id"].(string)
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"repository": inst.instance.Describe(),
		"items":      items,
	})
}

func handleListVersions(w http.ResponseWriter, r *http.Request) {
	itemID := r.PathValue("itemId")
	repoID, host, name, err := parseItemID(itemID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
		return
	}
	inst := getInstanceByRepoID(repoID)
	if inst == nil || inst.common == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found")
		return
	}
	ctx := r.Context()
	labelBindings, err := inst.common.GetLabels(ctx, Locator{Host: host, Name: name})
	if err != nil && !isNotFoundError(err) {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	labelToID := map[string]string{}
	if labelBindings != nil {
		for k, v := range labelBindings.Labels {
			labelToID[k] = v
		}
	}
	summaries, err := inst.common.ListVersions(ctx, Locator{Host: host, Name: name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	var versions []map[string]any
	seenLabels := map[string]struct{}{}
	for label, versionID := range labelToID {
		vm, err := inst.common.GetVersionMeta(ctx, Locator{Host: host, Name: name, VersionID: versionID})
		if err != nil && !isNotFoundError(err) {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
			return
		}
		versions = append(versions, map[string]any{
			"id":     repoID + ":" + host + "/" + name + ":" + label,
			"label":  label,
			"detail": formatVersionDetail(vm),
		})
		seenLabels[label] = struct{}{}
	}
	for _, s := range summaries {
		// If there was no label for this version, expose the UUID as the label.
		uuidLabel := s.VersionID
		if _, ok := seenLabels[uuidLabel]; ok {
			continue
		}
		versions = append(versions, map[string]any{
			"id":     repoID + ":" + host + "/" + name + ":" + uuidLabel,
			"label":  uuidLabel,
			"detail": formatVersionDetail(nil),
		})
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i]["label"].(string) < versions[j]["label"].(string)
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"item": map[string]any{
			"id":    repoID + ":" + host + "/" + name,
			"label": name,
		},
		"versions": versions,
	})
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("versionId")
	repoID, host, name, versionLabel, err := parseVersionID(versionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
		return
	}
	inst := getInstanceByRepoID(repoID)
	if inst == nil || inst.common == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found")
		return
	}
	ctx := r.Context()
	meta, resolvedLabel, err := resolveVersionMeta(ctx, inst.common, host, name, versionLabel)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	files := make([]map[string]any, 0, len(meta.Files))
	for _, f := range meta.Files {
		files = append(files, map[string]any{
			"id":           repoID + ":" + host + "/" + name + ":" + resolvedLabel + ":" + f.Name,
			"name":         f.Name,
			"path":         host + "/" + name + "/" + resolvedLabel + "/" + f.Name,
			"content_type": f.MediaType,
			"size_bytes":   f.Size,
			"modified":     meta.CreatedAt.Format(time.RFC3339),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i]["name"].(string) < files[j]["name"].(string) })
	writeJSON(w, http.StatusOK, map[string]any{
		"version": map[string]any{
			"id":    repoID + ":" + host + "/" + name + ":" + resolvedLabel,
			"label": resolvedLabel,
		},
		"files": files,
	})
}

func handleFileDetail(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	repoID, host, name, versionLabel, fileName, err := parseFileID(fileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
		return
	}
	inst := getInstanceByRepoID(repoID)
	if inst == nil || inst.common == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found")
		return
	}
	ctx := r.Context()
	meta, resolvedLabel, err := resolveVersionMeta(ctx, inst.common, host, name, versionLabel)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	var found *FileEntry
	for _, f := range meta.Files {
		if f.Name == fileName {
			found = &f
			break
		}
	}
	if found == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "file not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"file": map[string]any{
			"id":              repoID + ":" + host + "/" + name + ":" + resolvedLabel + ":" + found.Name,
			"name":            found.Name,
			"path":            host + "/" + name + "/" + resolvedLabel + "/" + found.Name,
			"repository_type": inst.instance.Describe().TypeID,
			"repository_name": inst.instance.Describe().Label,
			"item_label":      name,
			"version_label":   resolvedLabel,
			"content_type":    found.MediaType,
			"size_bytes":      found.Size,
			"checksum": map[string]any{
				"algorithm": strings.SplitN(found.BlobKey, ":", 2)[0],
				"value":     strings.SplitN(found.BlobKey, ":", 2)[1],
			},
			"download_count": nil,
			"last_accessed":  nil,
		},
	})
}

func resolveVersionMeta(ctx context.Context, store CommonStorage, host, name, versionLabel string) (*VersionMeta, string, error) {
	// First try label resolution.
	if versionLabel != "" {
		loc, err := store.ResolveLabel(ctx, Locator{
			Host:  host,
			Name:  name,
			Label: versionLabel,
		})
		if err == nil && loc.VersionID != "" {
			meta, err := store.GetVersionMeta(ctx, loc)
			return meta, versionLabel, err
		}
	}
	// Fallback to treating versionLabel as a direct version ID.
	loc := Locator{Host: host, Name: name, VersionID: versionLabel}
	meta, err := store.GetVersionMeta(ctx, loc)
	if err != nil {
		return nil, "", err
	}
	return meta, versionLabel, nil
}

func parseItemID(id string) (repoID, host, name string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid item id")
	}
	repoID = parts[0]
	hostName := parts[1]
	hostParts := strings.SplitN(hostName, "/", 2)
	if len(hostParts) != 2 {
		return "", "", "", fmt.Errorf("invalid item id")
	}
	return repoID, hostParts[0], hostParts[1], nil
}

func parseVersionID(id string) (repoID, host, name, versionLabel string, err error) {
	firstSplit := strings.SplitN(id, ":", 2)
	if len(firstSplit) != 2 {
		return "", "", "", "", fmt.Errorf("invalid version id")
	}
	repoID = firstSplit[0]
	rest := firstSplit[1]
	hostLabelSplit := strings.SplitN(rest, ":", 2)
	if len(hostLabelSplit) != 2 {
		return "", "", "", "", fmt.Errorf("invalid version id")
	}
	hostName := hostLabelSplit[0]
	versionLabel = hostLabelSplit[1]
	hostParts := strings.SplitN(hostName, "/", 2)
	if len(hostParts) != 2 {
		return "", "", "", "", fmt.Errorf("invalid version id")
	}
	host = hostParts[0]
	name = hostParts[1]
	return
}

func parseFileID(id string) (repoID, host, name, versionLabel, fileName string, err error) {
	firstSplit := strings.SplitN(id, ":", 2)
	if len(firstSplit) != 2 {
		return "", "", "", "", "", fmt.Errorf("invalid file id")
	}
	repoID = firstSplit[0]
	rest := firstSplit[1]

	// fileName is after the last colon to allow colons inside version labels.
	lastColon := strings.LastIndex(rest, ":")
	if lastColon <= 0 || lastColon == len(rest)-1 {
		return "", "", "", "", "", fmt.Errorf("invalid file id")
	}
	fileName = rest[lastColon+1:]
	hostAndVersion := rest[:lastColon]

	hostSplit := strings.SplitN(hostAndVersion, ":", 2)
	if len(hostSplit) != 2 {
		return "", "", "", "", "", fmt.Errorf("invalid file id")
	}
	hostName := hostSplit[0]
	versionLabel = hostSplit[1]

	hostParts := strings.SplitN(hostName, "/", 2)
	if len(hostParts) != 2 {
		return "", "", "", "", "", fmt.Errorf("invalid file id")
	}
	host = hostParts[0]
	name = hostParts[1]
	return
}

func getTypeByID(typeID string) *TypeDetails {
	rTypeLock.RLock()
	defer rTypeLock.RUnlock()
	for _, td := range rTypeDetails {
		meta := td.rType.Meta()
		metaID := meta.ID
		if metaID == "" {
			metaID = td.typeName
		}
		if metaID == typeID || td.typeName == typeID {
			return td
		}
	}
	return nil
}

func getInstanceByRepoID(repoID string) *InstanceDetails {
	rTypeLock.RLock()
	defer rTypeLock.RUnlock()
	for _, td := range rTypeDetails {
		if td == nil {
			continue
		}
		if inst, ok := td.instances[repoID]; ok {
			return inst
		}
	}
	return nil
}

func formatVersionDetail(meta *VersionMeta) string {
	if meta == nil || meta.CreatedAt.IsZero() {
		return ""
	}
	return fmt.Sprintf("Published %s", meta.CreatedAt.Format("2006-01-02"))
}
