import type React from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import './app.css';
import type {
  PanelDescriptor,
  BreadcrumbItem,
  RepoType,
  FileRow,
  ToastMessage,
  RepoItem,
  RepoItemVersion,
} from './components';
import { HeaderBar, ConcertinaShell, ToastQueue, SettingsDialog, FloatingDebugButton } from './components';
import { RepositoryTypesPanel, RepoInstancesPanel, ItemListPanel, VersionListPanel, FileListPanel, FileDetailsPanel } from './panels';
import {
  sampleData,
  type SampleRepoType,
  type SampleFile,
  type SampleNameEntry,
  type SampleRepo,
} from './mock/sampleData';
import {
  loadRepositoryTypes,
  loadRepositories,
  loadItems,
  loadVersions,
  loadFiles,
  loadFileDetail,
  type ApiType,
  type ApiRepo,
  type ApiItem,
  type ApiVersion,
  type ApiFile,
  type ApiFileDetail,
} from './api/uiClient';

interface RepoTypeMeta {
  id: string;
  label: string;
  description?: string;
}

interface RepoSummary {
  id: string;
  label: string;
  description?: string;
  typeId: string;
}

interface BuiltRepoData {
  items: RepoItem[];
  versionsByItemId: Record<string, RepoItemVersion[]>;
  versionsById: Record<string, RepoItemVersion>;
  filesByVersionId: Record<string, FileRow[]>;
  fileMeta: Record<string, FileDetail>;
}

interface FileDetail {
  id: string;
  name: string;
  path: string;
  repoType: string;
  repoName: string;
  sizeBytes: number;
  modified: string;
  contentType: string;
  checksum: string;
  downloadCount: number | null;
  lastAccessed: string | null;
}

interface DataCache {
  repoTypes: RepoTypeMeta[];
  reposByType: Record<string, RepoSummary[]>;
  itemsByRepo: Record<string, RepoItem[]>;
  versionsByItemId: Record<string, RepoItemVersion[]>;
  versionsById: Record<string, RepoItemVersion>;
  filesByVersionId: Record<string, FileRow[]>;
  fileMeta: Record<string, FileDetail>;
}

const systemPrefersDark =
  typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches;

export default function App() {
  const stored = loadStoredSettings();
  const [themeMode, setThemeMode] = useState<'light' | 'dark' | 'system'>(stored.themeMode ?? 'light');
  const [density, setDensity] = useState<'comfortable' | 'compact'>(stored.density ?? 'comfortable');
  const [dataSource, setDataSource] = useState<'simulated' | 'backend'>(stored.dataSource ?? 'backend');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [selectedTypeId, setSelectedTypeId] = useState<string | null>(null);
  const [selectedRepoId, setSelectedRepoId] = useState<string | null>(null);
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null);
  const [selectedVersionId, setSelectedVersionId] = useState<string | null>(null);
  const [selectedFileId, setSelectedFileId] = useState<string | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>(
    (stored.dataSource ?? 'backend') === 'simulated'
      ? sampleData.toasts
      : [{ id: 'backend', level: 'success', message: 'Connected to backend API' }],
  );
  const hasAppliedDataSourceReset = useRef(false);
  const isHydratingFromURL = useRef(true);
  const [dataCache, setDataCache] = useState<DataCache>(() =>
    (stored.dataSource ?? 'backend') === 'simulated' ? buildSimulatedCache() : emptyCache(),
  );
  const [repoTypesLoading, setRepoTypesLoading] = useState(false);
  const [repoTypesError, setRepoTypesError] = useState<string | null>(null);
  const [reposLoading, setReposLoading] = useState(false);
  const [reposError, setReposError] = useState<string | null>(null);
  const [itemsLoading, setItemsLoading] = useState(false);
  const [itemsError, setItemsError] = useState<string | null>(null);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [versionsError, setVersionsError] = useState<string | null>(null);
  const [filesLoading, setFilesLoading] = useState(false);
  const [filesError, setFilesError] = useState<string | null>(null);
  const [fileDetailError, setFileDetailError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => {
    if (typeof window === 'undefined') {
      return '/api/ui/v1/';
    }
    const { protocol, host } = window.location;
    return `${protocol}//${host}/api/ui/v1/`;
  }, []);

  const uiBasePath = useMemo(() => {
    if (typeof window === 'undefined') return '/ui/';
    const path = window.location.pathname;
    const marker = '/ui/';
    const idx = path.indexOf(marker);
    if (idx >= 0) {
      return path.slice(0, idx + marker.length);
    }
    return path.endsWith('/') ? path : `${path}/`;
  }, []);

  // Apply route from URL on first load and when the user navigates with back/forward.
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const applyRoute = (route: RouteSelection) => {
      console.log('[route] applyRoute', { uiBasePath, route });
      setSelectedTypeId(route.typeId ?? null);
      setSelectedRepoId(route.repoId ?? null);
      setSelectedItemId(route.itemId ?? null);
      setSelectedVersionId(route.versionId ?? null);
      setSelectedFileId(route.fileId ?? null);
    };
    applyRoute(parseRoute(window.location.pathname, uiBasePath));
    const onPop = () => applyRoute(parseRoute(window.location.pathname, uiBasePath));
    window.addEventListener('popstate', onPop);
    isHydratingFromURL.current = false;
    return () => window.removeEventListener('popstate', onPop);
  }, [uiBasePath]);

  // Reset data & selections when switching data source.
  useEffect(() => {
    if (!hasAppliedDataSourceReset.current) {
      hasAppliedDataSourceReset.current = true;
      return;
    }
    setSelectedTypeId(null);
    setSelectedRepoId(null);
    setSelectedItemId(null);
    setSelectedVersionId(null);
    setSelectedFileId(null);
    setRepoTypesError(null);
    setReposError(null);
    setItemsError(null);
    setVersionsError(null);
    setFilesError(null);
    setFileDetailError(null);
    if (dataSource === 'simulated') {
      setDataCache(buildSimulatedCache());
      setToasts(sampleData.toasts);
    } else {
      setDataCache(emptyCache());
      setToasts([{ id: 'backend', level: 'success', message: 'Connected to backend API' }]);
    }
  }, [dataSource]);

  // Persist settings to storage/cookie.
  useEffect(() => {
    persistSettings({ themeMode, density, dataSource });
  }, [themeMode, density, dataSource]);

  // Backend: load repository types.
  useEffect(() => {
    if (dataSource !== 'backend') {
      setRepoTypesLoading(false);
      return;
    }
    let cancelled = false;
    setRepoTypesLoading(true);
    setRepoTypesError(null);
    loadRepositoryTypes(apiBaseUrl)
      .then((types) => {
        if (cancelled) return;
        setDataCache((prev) => ({
          ...prev,
          repoTypes: types.map(mapApiType),
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setRepoTypesError(err.message);
      })
      .finally(() => {
        if (!cancelled) {
          setRepoTypesLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, apiBaseUrl]);

  // Backend: load repositories for selected type.
  useEffect(() => {
    if (dataSource !== 'backend' || !selectedTypeId || dataCache.reposByType[selectedTypeId]) {
      return;
    }
    let cancelled = false;
    setReposLoading(true);
    setReposError(null);
    loadRepositories(apiBaseUrl, selectedTypeId)
      .then((repos) => {
        if (cancelled) return;
        setDataCache((prev) => ({
          ...prev,
          reposByType: {
            ...prev.reposByType,
            [selectedTypeId]: repos.map((repo) => mapApiRepo(repo, selectedTypeId)),
          },
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setReposError(err.message);
      })
      .finally(() => {
        if (!cancelled) {
          setReposLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, selectedTypeId, apiBaseUrl, dataCache.reposByType]);

  // Backend: load items for selected repo.
  useEffect(() => {
    if (dataSource !== 'backend' || !selectedRepoId || dataCache.itemsByRepo[selectedRepoId]) {
      return;
    }
    let cancelled = false;
    setItemsLoading(true);
    setItemsError(null);
    loadItems(apiBaseUrl, selectedRepoId)
      .then((items) => {
        if (cancelled) return;
        setDataCache((prev) => ({
          ...prev,
          itemsByRepo: {
            ...prev.itemsByRepo,
            [selectedRepoId]: items.map(mapApiItem),
          },
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setItemsError(err.message);
      })
      .finally(() => {
        if (!cancelled) {
          setItemsLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, selectedRepoId, apiBaseUrl, dataCache.itemsByRepo]);

  // Backend: load versions for selected item.
  useEffect(() => {
    if (dataSource !== 'backend' || !selectedItemId || dataCache.versionsByItemId[selectedItemId]) {
      return;
    }
    let cancelled = false;
    setVersionsLoading(true);
    setVersionsError(null);
    loadVersions(apiBaseUrl, selectedItemId)
      .then((versions) => {
        if (cancelled) return;
        const mappedVersions = versions.map((version) => mapApiVersion(version, selectedItemId));
        const mappedById: Record<string, RepoItemVersion> = {};
        mappedVersions.forEach((v) => {
          mappedById[v.id] = v;
        });
        setDataCache((prev) => ({
          ...prev,
          versionsByItemId: {
            ...prev.versionsByItemId,
            [selectedItemId]: mappedVersions,
          },
          versionsById: {
            ...prev.versionsById,
            ...mappedById,
          },
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setVersionsError(err.message);
      })
      .finally(() => {
        if (!cancelled) {
          setVersionsLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, selectedItemId, apiBaseUrl, dataCache.versionsByItemId]);

  // Backend: load files for selected version.
  useEffect(() => {
    if (dataSource !== 'backend' || !selectedVersionId || dataCache.filesByVersionId[selectedVersionId]) {
      return;
    }
    let cancelled = false;
    setFilesLoading(true);
    setFilesError(null);
    loadFiles(apiBaseUrl, selectedVersionId)
      .then((files) => {
        if (cancelled) return;
        setDataCache((prev) => ({
          ...prev,
          filesByVersionId: {
            ...prev.filesByVersionId,
            [selectedVersionId]: files.map(mapApiFile),
          },
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setFilesError(err.message);
      })
      .finally(() => {
        if (!cancelled) {
          setFilesLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, selectedVersionId, apiBaseUrl, dataCache.filesByVersionId]);

  // Backend: load file detail when a file is selected.
  useEffect(() => {
    if (dataSource !== 'backend' || !selectedFileId || dataCache.fileMeta[selectedFileId]) {
      return;
    }
    let cancelled = false;
    setFileDetailError(null);
    loadFileDetail(apiBaseUrl, selectedFileId)
      .then((detail) => {
        if (cancelled) return;
        setDataCache((prev) => ({
          ...prev,
          fileMeta: {
            ...prev.fileMeta,
            [detail.id]: mapApiFileDetail(detail),
          },
        }));
      })
      .catch((err: Error) => {
        if (cancelled) return;
        setFileDetailError(err.message);
      });
    return () => {
      cancelled = true;
    };
  }, [dataSource, selectedFileId, apiBaseUrl, dataCache.fileMeta]);

  const repoTypeTiles: RepoType[] = dataCache.repoTypes.map((type) => ({
    id: type.id,
    label: type.label,
    description: type.description,
    onSelect: () => {
      setSelectedTypeId(type.id);
      setSelectedRepoId(null);
      setSelectedItemId(null);
      setSelectedVersionId(null);
      setSelectedFileId(null);
    },
  }));

  const selectedType = selectedTypeId
    ? dataCache.repoTypes.find((type) => type.id === selectedTypeId) ?? null
    : null;
  const reposForType = selectedTypeId ? dataCache.reposByType[selectedTypeId] ?? [] : [];
  const selectedRepo = selectedRepoId ? reposForType.find((repo) => repo.id === selectedRepoId) ?? null : null;
  const items = selectedRepoId ? dataCache.itemsByRepo[selectedRepoId] ?? [] : [];
  const selectedItem = selectedItemId ? items.find((item) => item.id === selectedItemId) ?? null : null;
  const versionsForItem = selectedItemId ? dataCache.versionsByItemId[selectedItemId] ?? [] : [];
  const selectedVersion = selectedVersionId ? dataCache.versionsById[selectedVersionId] ?? null : null;
  const versionFiles = selectedVersionId ? dataCache.filesByVersionId[selectedVersionId] ?? [] : [];
  const selectedFile = selectedFileId ? dataCache.fileMeta[selectedFileId] ?? null : null;

  // Sync URL with current selection.
  useEffect(() => {
    if (typeof window === 'undefined') return;
    console.log('[route] sync selection -> url', {
      selectedTypeId,
      selectedRepoId,
      selectedItemId,
      selectedVersionId,
      selectedFileId,
    });
    const currentRoute = parseRoute(window.location.pathname, uiBasePath);
    if (
      (!selectedTypeId && currentRoute.typeId) ||
      (!selectedRepoId && currentRoute.repoId) ||
      (!selectedItemId && currentRoute.itemId) ||
      (!selectedVersionId && currentRoute.versionId) ||
      (!selectedFileId && currentRoute.fileId)
    ) {
      if (isHydratingFromURL.current) {
        // Don't collapse a deeper URL to root while state is still catching up.
        return;
      }
    }
    const route: RouteSelection = {
      typeId: selectedTypeId ?? undefined,
      repoId: selectedRepoId ?? undefined,
      itemId: selectedItemId ?? undefined,
      versionId: selectedVersionId ?? undefined,
      fileId: selectedFileId ?? undefined,
    };
    const targetPath = buildRoutePath(route, uiBasePath);
    if (window.location.pathname !== targetPath) {
      console.log('[route] pushState', { targetPath });
      window.history.pushState({}, '', targetPath + window.location.search + window.location.hash);
    }
  }, [selectedTypeId, selectedRepoId, selectedItemId, selectedVersionId, selectedFileId, uiBasePath]);

  const panels: PanelDescriptor[] = [];

  if (!selectedType) {
    panels.push({
      id: 'repository-types',
      title: 'Repoxy',
      content: (
        <RepositoryTypesPanel
          repoTypes={repoTypeTiles}
          selectedId={selectedTypeId}
          loading={dataSource === 'backend' ? repoTypesLoading : false}
          error={repoTypesError ?? undefined}
          emptyMessage="No repository types found."
        />
      ),
      mobileVisible: true,
    });
  } else if (!selectedRepo) {
    const repoTiles: RepoType[] = reposForType.map((repo) => ({
      id: repo.id,
      label: repo.label,
      description: repo.description ?? 'Routes requests through Repoxy for this repository.',
      onSelect: () => {
        setSelectedRepoId(repo.id);
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
      },
    }));

    let repoContent: React.ReactNode = (
      <RepoInstancesPanel repos={repoTiles} selectedId={selectedRepoId} emptyMessage="No repositories found." />
    );
    if (dataSource === 'backend' && reposLoading) {
      repoContent = <Alert severity="info">Loading repositories…</Alert>;
    } else if (reposError) {
      repoContent = <Alert severity="error">{reposError}</Alert>;
    }

    panels.push({
      id: `repos-${selectedType.id}`,
      title: `${selectedType.label} Repositories`,
      content: repoContent,
      mobileVisible: true,
    });
  } else {
    panels.push({
      id: `items-${selectedRepo.id}`,
      title: `${selectedRepo.label} Items`,
      content: (
        <ItemListPanel
          items={items}
          selectedItemId={selectedItemId}
          onItemSelect={(item) => {
            setSelectedItemId(item.id);
            setSelectedVersionId(null);
            setSelectedFileId(null);
          }}
          loadingMessage={dataSource === 'backend' && itemsLoading ? 'Loading items…' : undefined}
          emptyMessage={itemsError ?? 'No items found.'}
        />
      ),
      mobileVisible: !selectedItemId,
    });

    if (selectedItemId) {
      let versionsContent: React.ReactNode = (
        <VersionListPanel
          versions={versionsForItem}
          selectedVersionId={selectedVersionId}
          onVersionSelect={(version) => {
            setSelectedVersionId(version.id);
            setSelectedFileId(null);
          }}
          emptyMessage={versionsError ?? 'No versions found.'}
        />
      );
      if (dataSource === 'backend' && versionsLoading) {
        versionsContent = <Alert severity="info">Loading versions…</Alert>;
      }

      panels.push({
        id: `versions-${selectedItemId}`,
        title: `${selectedRepo.label} Versions`,
        content: versionsContent,
        mobileVisible: true,
      });

      if (selectedVersionId) {
        panels.push({
          id: `files-${selectedVersionId}`,
          title: `${selectedRepo.label} Files`,
          content: (
            <FileListPanel
              files={versionFiles}
              selectedFileId={selectedFileId}
              onFileSelect={(row) => setSelectedFileId(row.id)}
              filesLoading={dataSource === 'backend' ? filesLoading : false}
              filesError={filesError ?? undefined}
              emptyFilesMessage={
                selectedVersionId ? 'No files for this version.' : 'Select a version to view files.'
              }
            />
          ),
          mobileVisible: true,
        });
      }
    }
  }

  if (selectedFile) {
    const fileDetails = (
      <FileDetailsPanel
        title={selectedFile.name}
        subtitle={`${selectedFile.repoType} • ${selectedFile.repoName}`}
        metadata={[
          { label: 'Path', value: selectedFile.path },
          { label: 'Modified', value: selectedFile.modified || '—' },
          { label: 'Size', value: formatBytes(selectedFile.sizeBytes) },
          { label: 'Content Type', value: selectedFile.contentType },
          { label: 'Checksum', value: selectedFile.checksum },
        ]}
        usage={[
          { label: 'Download Count', value: selectedFile.downloadCount ?? '—' },
          { label: 'Last Accessed', value: selectedFile.lastAccessed ?? '—' },
        ]}
      />
    );

    panels.push({
      id: `file-${selectedFile.id}`,
      title: selectedFile.name,
      content: fileDetailError ? <Alert severity="error">{fileDetailError}</Alert> : fileDetails,
      mobileVisible: true,
    });
  }

  const breadcrumbs: BreadcrumbItem[] = [
    {
      id: 'crumb-types',
      label: 'Repoxy',
      onSelect: () => {
        setSelectedTypeId(null);
        setSelectedRepoId(null);
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
        pushRoute({}, uiBasePath);
      },
      isCurrent: !selectedType,
    },
  ];

  if (selectedType) {
    breadcrumbs.push({
      id: `crumb-${selectedType.id}`,
      label: selectedType.label,
      onSelect: () => {
        setSelectedRepoId(null);
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
        pushRoute({ typeId: selectedType.id }, uiBasePath);
      },
      isCurrent: !selectedRepo,
    });
  }

  if (selectedType && selectedRepo) {
    breadcrumbs.push({
      id: `crumb-${selectedType.id}-${selectedRepo.id}`,
      label: selectedRepo.label,
      onSelect: () => {
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
        pushRoute({ typeId: selectedType.id, repoId: selectedRepo.id }, uiBasePath);
      },
      isCurrent: !selectedItem,
    });
  }

  if (selectedItem && selectedRepo) {
    breadcrumbs.push({
      id: `crumb-item-${selectedItem.id}`,
      label: selectedItem.label,
      onSelect: () => {
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
        pushRoute({ typeId: selectedType?.id, repoId: selectedRepo?.id }, uiBasePath);
      },
      isCurrent: !selectedVersion,
    });
  }

  if (selectedVersion) {
    breadcrumbs.push({
      id: `crumb-version-${selectedVersion.id}`,
      label: selectedVersion.label,
      onSelect: () => {
        setSelectedVersionId(null);
        setSelectedFileId(null);
        pushRoute({ typeId: selectedType?.id, repoId: selectedRepo?.id, itemId: selectedItem?.id }, uiBasePath);
      },
      isCurrent: !selectedFile,
    });
  }

  if (selectedFile) {
    breadcrumbs.push({
      id: `crumb-file-${selectedFile.id}`,
      label: selectedFile.name,
      isCurrent: true,
    });
  }

  const theme = useMemo(() => {
    let mode: 'light' | 'dark';
    if (themeMode === 'system') {
      mode = systemPrefersDark ? 'dark' : 'light';
    } else {
      mode = themeMode;
    }
    return createTheme({ palette: { mode } });
  }, [themeMode]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box
        className={`app-shell density-${density}`}
        display="flex"
        flexDirection="column"
        minHeight="100vh"
        height="100vh"
        sx={{ backgroundColor: (theme) => theme.palette.background.default, overflow: 'hidden' }}
      >
        <HeaderBar breadcrumbs={breadcrumbs} onOpenSettings={() => setSettingsOpen(true)} />
        <Toolbar />
        <Box
          component="main"
          className="app-main"
          flex={1}
          display="flex"
          flexDirection="column"
          overflow="hidden"
          sx={{ minHeight: 0 }}
        >
          <ConcertinaShell panels={panels} />
        </Box>
        <FloatingDebugButton />
        <SettingsDialog
          open={settingsOpen}
          density={density}
          themeMode={themeMode}
          dataSource={dataSource}
          apiBaseUrl={apiBaseUrl}
          onDensityChange={setDensity}
          onThemeChange={setThemeMode}
          onDataSourceChange={setDataSource}
          onClose={() => setSettingsOpen(false)}
        />
        <ToastQueue toasts={toasts} onDismiss={(id) => setToasts((current) => current.filter((toast) => toast.id !== id))} />
      </Box>
    </ThemeProvider>
  );
}

function emptyCache(): DataCache {
  return {
    repoTypes: [],
    reposByType: {},
    itemsByRepo: {},
    versionsByItemId: {},
    versionsById: {},
    filesByVersionId: {},
    fileMeta: {},
  };
}

function buildRepoData(repoType: SampleRepoType, repo: SampleRepo): BuiltRepoData {
  const items: RepoItem[] = [];
  const versionsByItemId: Record<string, RepoItemVersion[]> = {};
  const versionsById: Record<string, RepoItemVersion> = {};
  const filesByVersionId: Record<string, FileRow[]> = {};
  const fileMeta: Record<string, FileDetail> = {};

  for (const host of repo.hosts) {
    for (const group of host.groups) {
      for (const nameEntry of group.names) {
        const itemId = `item:${repoType.type}:${repo.name}:${host.host}:${group.group}:${nameEntry.name}`;
        const item: RepoItem = {
          id: itemId,
          label: `${group.group}/${nameEntry.name}`,
          description: host.host,
          path: [repo.display_name, host.host, group.group, nameEntry.name],
        };
        items.push(item);

        const versionId = `${itemId}:latest`;
        const versionLabel = nameEntry.last_updated
          ? `Latest • ${new Date(nameEntry.last_updated).toLocaleDateString()}`
          : 'Latest';
        const version: RepoItemVersion = {
          id: versionId,
          itemId,
          label: versionLabel,
          description: `Updated ${nameEntry.last_updated ?? 'recently'}`,
        };
        versionsByItemId[itemId] = [version];
        versionsById[versionId] = version;

        const fileRows = nameEntry.files.map((file) => {
          const row = transformFile(repoType, repo, nameEntry, file);
          const detail = buildFileDetail(repoType, repo, nameEntry, file);
          fileMeta[detail.id] = detail;
          return row;
        });
        filesByVersionId[versionId] = fileRows;
      }
    }
  }

  return { items, versionsByItemId, versionsById, filesByVersionId, fileMeta };
}

function transformFile(repoType: SampleRepoType, repo: SampleRepo, nameEntry: SampleNameEntry, file: SampleFile): FileRow {
  const id = `${repoType.type}:${repo.name}:${nameEntry.name}:${file.file}`;
  return {
    id,
    name: file.file,
    modified: file.modified,
    sizeBytes: file.size_bytes,
    path: file.path,
  };
}

function buildFileDetail(repoType: SampleRepoType, repo: SampleRepo, nameEntry: SampleNameEntry, file: SampleFile): FileDetail {
  const id = `${repoType.type}:${repo.name}:${nameEntry.name}:${file.file}`;
  return {
    id,
    name: file.file,
    path: file.path,
    repoType: repoType.label,
    repoName: repo.display_name,
    sizeBytes: file.size_bytes,
    modified: file.modified,
    contentType: file.content_type,
    checksum: file.checksums.sha256,
    downloadCount: file.download_count,
    lastAccessed: file.last_accessed,
  };
}

function formatBytes(size: number) {
  if (size < 1024) return `${size} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let value = size / 1024;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`;
}

type RouteSelection = {
  typeId?: string;
  repoId?: string;
  itemId?: string;
  versionId?: string;
  fileId?: string;
};

type StoredSettings = {
  themeMode?: 'light' | 'dark' | 'system';
  density?: 'comfortable' | 'compact';
  dataSource?: 'simulated' | 'backend';
};

const SETTINGS_KEY = 'repoxy_ui_settings';

function loadStoredSettings(): StoredSettings {
  if (typeof window === 'undefined') return {};
  try {
    const fromLocal = window.localStorage.getItem(SETTINGS_KEY);
    if (fromLocal) {
      return JSON.parse(fromLocal) as StoredSettings;
    }
  } catch {
    // ignore
  }
  const fromCookie = readCookie(SETTINGS_KEY);
  if (fromCookie) {
    try {
      return JSON.parse(fromCookie) as StoredSettings;
    } catch {
      return {};
    }
  }
  return {};
}

function persistSettings(settings: StoredSettings) {
  if (typeof window === 'undefined') return;
  const value = JSON.stringify(settings);
  try {
    window.localStorage.setItem(SETTINGS_KEY, value);
  } catch {
    // ignore
  }
  setCookie(SETTINGS_KEY, value, 365);
}

function setCookie(name: string, value: string, days: number) {
  if (typeof document === 'undefined') return;
  const expires = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString();
  document.cookie = `${encodeURIComponent(name)}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Lax`;
}

function readCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;
  const cookies = document.cookie ? document.cookie.split('; ') : [];
  for (const c of cookies) {
    const [key, ...rest] = c.split('=');
    if (decodeURIComponent(key) === name) {
      return decodeURIComponent(rest.join('='));
    }
  }
  return null;
}

function parseRoute(pathname: string, basePath: string): RouteSelection {
  const normalisedBase = basePath.endsWith('/') ? basePath : `${basePath}/`;
  if (!pathname.startsWith(normalisedBase)) {
    console.log('[route] parseRoute: outside base', { pathname, basePath });
    return {};
  }
  const rel = pathname.slice(normalisedBase.length);
  const parts = rel.split('/').filter(Boolean).map(decodeURIComponent);
  const [typeId, repoId, itemId, versionId, fileId] = parts;
  const route: RouteSelection = {};
  if (typeId) route.typeId = typeId;
  if (repoId) route.repoId = repoId;
  if (itemId) route.itemId = itemId;
  if (versionId) route.versionId = versionId;
  if (fileId) route.fileId = fileId;
  return route;
}

function buildRoutePath(route: RouteSelection, basePath: string): string {
  const segments: string[] = [];
  if (route.typeId) {
    segments.push(encodeURIComponent(route.typeId));
    if (route.repoId) {
      segments.push(encodeURIComponent(route.repoId));
      if (route.itemId) {
        segments.push(encodeURIComponent(route.itemId));
        if (route.versionId) {
          segments.push(encodeURIComponent(route.versionId));
          if (route.fileId) {
            segments.push(encodeURIComponent(route.fileId));
          }
        }
      }
    }
  }
  const base = basePath.endsWith('/') ? basePath.slice(0, -1) : basePath;
  const suffix = segments.length ? `/${segments.join('/')}` : '';
  return `${base}${suffix}/`;
}

function buildSimulatedCache(): DataCache {
  const cache = emptyCache();

  for (const repoType of sampleData.repository_types) {
    cache.repoTypes.push({
      id: repoType.type,
      label: repoType.label,
      description: repoType.description,
    });

    cache.reposByType[repoType.type] = repoType.repos.map((repo) => ({
      id: repo.name,
      label: repo.display_name,
      description: 'Routes requests through Repoxy for this repository.',
      typeId: repoType.type,
    }));

    for (const repo of repoType.repos) {
      const built = buildRepoData(repoType, repo);
      cache.itemsByRepo[repo.name] = built.items;
      cache.versionsByItemId = { ...cache.versionsByItemId, ...built.versionsByItemId };
      cache.versionsById = { ...cache.versionsById, ...built.versionsById };
      cache.filesByVersionId = { ...cache.filesByVersionId, ...built.filesByVersionId };
      cache.fileMeta = { ...cache.fileMeta, ...built.fileMeta };
    }
  }

  return cache;
}

function parseItemPath(id: string): string[] {
  const parts = id.split(':');
  if (parts.length < 2) {
    return [id];
  }
  const hostAndName = parts[1];
  const hostParts = hostAndName.split('/');
  return hostParts.length === 2 ? hostParts : [hostAndName];
}

function mapApiType(t: ApiType): RepoTypeMeta {
  const id = t.id ?? '';
  return {
    id,
    label: (t.label ?? id) || 'Repository Type',
    description: t.description,
  };
}

function mapApiRepo(repo: ApiRepo, typeId: string): RepoSummary {
  const id = repo.id ?? '';
  return {
    id,
    label: (repo.label ?? id) || 'Repository',
    description: repo.description,
    typeId: typeId || repo.type_id || '',
  };
}

function mapApiItem(item: ApiItem): RepoItem {
  return {
    id: item.id,
    label: item.label ?? item.id,
    description: item.detail,
    path: parseItemPath(item.id),
  };
}

function mapApiVersion(version: ApiVersion, itemId: string): RepoItemVersion {
  return {
    id: version.id,
    itemId,
    label: version.label ?? version.id,
    description: version.detail,
  };
}

function mapApiFile(file: ApiFile): FileRow {
  return {
    id: file.id,
    name: file.name,
    modified: file.modified ?? '',
    sizeBytes: file.size_bytes ?? 0,
    path: file.path,
  };
}

function mapApiFileDetail(file: ApiFileDetail): FileDetail {
  return {
    id: file.id,
    name: file.name,
    path: file.path ?? '',
    repoType: file.repository_type ?? '',
    repoName: file.repository_name ?? '',
    sizeBytes: file.size_bytes ?? 0,
    modified: file.last_accessed ?? '',
    contentType: file.content_type ?? '',
    checksum: file.checksum?.value ?? '',
    downloadCount: file.download_count ?? null,
    lastAccessed: file.last_accessed ?? null,
  };
}
