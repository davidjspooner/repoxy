import { useMemo, useState } from 'react';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, ThemeProvider } from '@mui/material/styles';
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
  downloadCount: number;
  lastAccessed: string;
}

const systemPrefersDark =
  typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches;

export default function App() {
  const [themeMode, setThemeMode] = useState<'light' | 'dark' | 'system'>('light');
  const [density, setDensity] = useState<'comfortable' | 'compact'>('comfortable');
  const [dataSource, setDataSource] = useState<'simulated' | 'backend'>('simulated');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [selectedTypeId, setSelectedTypeId] = useState<string | null>(null);
  const [selectedRepoId, setSelectedRepoId] = useState<string | null>(null);
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null);
  const [selectedVersionId, setSelectedVersionId] = useState<string | null>(null);
  const [selectedFileId, setSelectedFileId] = useState<string | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>(sampleData.toasts);

  const apiBaseUrl = useMemo(() => {
    if (typeof window === 'undefined') {
      return '/api/ui/v1/';
    }
    const { protocol, host } = window.location;
    return `${protocol}//${host}/api/ui/v1/`;
  }, []);

  const repoTypeMap = useMemo(() => new Map(sampleData.repository_types.map((type) => [type.type, type])), []);
  const selectedType = selectedTypeId ? repoTypeMap.get(selectedTypeId) ?? null : null;
  const selectedRepo =
    selectedType && selectedRepoId ? selectedType.repos.find((repo) => repo.name === selectedRepoId) ?? null : null;

  const builtRepo = useMemo(
    () => (selectedType && selectedRepo ? buildRepoData(selectedType, selectedRepo) : null),
    [selectedType, selectedRepo],
  );

  const items = builtRepo?.items ?? [];
  const selectedItem = selectedItemId ? items.find((item) => item.id === selectedItemId) ?? null : null;
  const versionsForItem =
    selectedItemId && builtRepo ? builtRepo.versionsByItemId[selectedItemId] ?? [] : [];
  const selectedVersion =
    selectedVersionId && builtRepo ? builtRepo.versionsById[selectedVersionId] ?? null : null;
  const versionFiles =
    selectedVersionId && builtRepo ? builtRepo.filesByVersionId[selectedVersionId] ?? [] : [];
  const selectedFile = selectedFileId && builtRepo ? builtRepo.fileMeta[selectedFileId] ?? null : null;

  const repoTypes: RepoType[] = sampleData.repository_types.map((type) => ({
    id: type.type,
    label: type.label,
    description: type.description,
    onSelect: () => {
      setSelectedTypeId(type.type);
      setSelectedRepoId(null);
      setSelectedItemId(null);
      setSelectedVersionId(null);
      setSelectedFileId(null);
    },
  }));

  const panels: PanelDescriptor[] = [];

  if (!selectedType) {
    panels.push({
      id: 'repository-types',
      title: 'Repoxy',
      content: <RepositoryTypesPanel repoTypes={repoTypes} selectedId={selectedTypeId} />,
      mobileVisible: true,
    });
  } else if (!selectedRepo) {
    const repoTiles: RepoType[] = selectedType.repos.map((repo) => ({
      id: repo.name,
      label: repo.display_name,
      description: 'Routes requests through Repoxy for this repository.',
      onSelect: () => {
        setSelectedRepoId(repo.name);
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
      },
    }));

    panels.push({
      id: `repos-${selectedType.type}`,
      title: `${selectedType.label} Repositories`,
      content: (
        <RepoInstancesPanel repos={repoTiles} selectedId={selectedRepoId} />
      ),
      mobileVisible: true,
    });
  } else if (builtRepo) {
    panels.push({
      id: `items-${selectedRepo.name}`,
      title: `${selectedRepo.display_name} Items`,
      content: (
        <ItemListPanel
          items={builtRepo.items}
          selectedItemId={selectedItemId}
          onItemSelect={(item) => {
            setSelectedItemId(item.id);
            setSelectedVersionId(null);
            setSelectedFileId(null);
          }}
        />
      ),
      mobileVisible: !selectedItemId,
    });

    if (selectedItemId) {
      panels.push({
        id: `versions-${selectedItemId}`,
        title: `${selectedRepo.display_name} Versions`,
        content: (
          <VersionListPanel
            versions={versionsForItem}
            selectedVersionId={selectedVersionId}
            onVersionSelect={(version) => {
              setSelectedVersionId(version.id);
              setSelectedFileId(null);
            }}
          />
        ),
        mobileVisible: true,
      });

      if (selectedVersionId) {
        panels.push({
          id: `files-${selectedVersionId}`,
          title: `${selectedRepo.display_name} Files`,
          content: (
            <FileListPanel
              files={versionFiles}
              selectedFileId={selectedFileId}
              onFileSelect={(row) => setSelectedFileId(row.id)}
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
    panels.push({
      id: `file-${selectedFile.id}`,
      title: selectedFile.name,
      content: (
        <FileDetailsPanel
          title={selectedFile.name}
          subtitle={`${selectedFile.repoType} • ${selectedFile.repoName}`}
          metadata={[
            { label: 'Path', value: selectedFile.path },
            { label: 'Modified', value: selectedFile.modified },
            { label: 'Size', value: formatBytes(selectedFile.sizeBytes) },
            { label: 'Content Type', value: selectedFile.contentType },
            { label: 'Checksum (SHA-256)', value: selectedFile.checksum },
          ]}
          usage={[
            { label: 'Download Count', value: selectedFile.downloadCount },
            { label: 'Last Accessed', value: selectedFile.lastAccessed },
          ]}
        />
      ),
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
      },
      isCurrent: !selectedType,
    },
  ];

  if (selectedType) {
    breadcrumbs.push({
      id: `crumb-${selectedType.type}`,
      label: selectedType.label,
      onSelect: () => {
        setSelectedRepoId(null);
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
      },
      isCurrent: !selectedRepo,
    });
  }

  if (selectedType && selectedRepo) {
    breadcrumbs.push({
      id: `crumb-${selectedType.type}-${selectedRepo.name}`,
      label: selectedRepo.display_name,
      onSelect: () => {
        setSelectedItemId(null);
        setSelectedVersionId(null);
        setSelectedFileId(null);
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
