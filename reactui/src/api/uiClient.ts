export interface ApiType {
  id?: string;
  label?: string;
  description?: string;
}

export interface ApiRepo {
  id?: string;
  label?: string;
  description?: string;
  type_id?: string;
}

export interface ApiItem {
  id: string;
  label: string;
  detail?: string;
}

export interface ApiVersion {
  id: string;
  label: string;
  detail?: string;
}

export interface ApiFile {
  id: string;
  name: string;
  path?: string;
  content_type?: string;
  size_bytes?: number;
  modified?: string;
}

export interface ApiFileDetail {
  id: string;
  name: string;
  path?: string;
  repository_type?: string;
  repository_name?: string;
  item_label?: string;
  version_label?: string;
  content_type?: string;
  size_bytes?: number;
  checksum?: {
    algorithm?: string;
    value?: string;
  };
  download_count?: number | null;
  last_accessed?: string | null;
}

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(url, { headers: { Accept: 'application/json' } });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Request failed (${response.status}): ${text || response.statusText}`);
  }
  return (await response.json()) as T;
}

export async function loadRepositoryTypes(apiBaseUrl: string): Promise<ApiType[]> {
  const data = await fetchJSON<{ types: ApiType[] }>(join(apiBaseUrl, 'repository-types'));
  return data.types ?? [];
}

export async function loadRepositories(apiBaseUrl: string, typeId: string): Promise<ApiRepo[]> {
  const data = await fetchJSON<{ repositories: ApiRepo[] }>(join(apiBaseUrl, `repository-types/${encodeURIComponent(typeId)}/repositories`));
  return data.repositories ?? [];
}

export async function loadItems(apiBaseUrl: string, repoId: string): Promise<ApiItem[]> {
  const data = await fetchJSON<{ items: ApiItem[] }>(
    join(apiBaseUrl, `repositories/${encodeURIComponent(repoId)}/items`),
  );
  return data.items ?? [];
}

export async function loadVersions(apiBaseUrl: string, itemId: string): Promise<ApiVersion[]> {
  const data = await fetchJSON<{ versions: ApiVersion[] }>(
    join(apiBaseUrl, `items/${encodeURIComponent(itemId)}/versions`),
  );
  return data.versions ?? [];
}

export async function loadFiles(apiBaseUrl: string, versionId: string): Promise<ApiFile[]> {
  const data = await fetchJSON<{ files: ApiFile[] }>(
    join(apiBaseUrl, `versions/${encodeURIComponent(versionId)}/files`),
  );
  return data.files ?? [];
}

export async function loadFileDetail(apiBaseUrl: string, fileId: string): Promise<ApiFileDetail> {
  const data = await fetchJSON<{ file: ApiFileDetail }>(join(apiBaseUrl, `files/${encodeURIComponent(fileId)}`));
  return data.file;
}

function join(base: string, path: string): string {
  if (!base.endsWith('/')) {
    base = `${base}/`;
  }
  return `${base}${path}`;
}
