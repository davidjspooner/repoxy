import { useMemo, useState } from 'react';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import CssBaseline from '@mui/material/CssBaseline';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import './app.css';
import type { PanelDescriptor, BreadcrumbItem, RepoType, FolderNode, FileRow, ToastMessage } from './components';
import { HeaderBar, ConcertinaShell, ToastQueue, SettingsDialog, FloatingDebugButton } from './components';
import { RepositoryTypesPanel, FolderBrowserPanel, FileListPanel, FileDetailsPanel } from './panels';
import {
  sampleData,
  type SampleRepoType,
  type SampleFile,
  type SampleNameEntry,
  type SampleRepo,
} from './mock/sampleData';

interface BuiltRepoData {
  folders: FolderNode[];
  filesByFolderId: Record<string, FileRow[]>;
  fileMeta: Record<string, FileDetail>;
  folderMeta: Record<string, FolderMeta>;
}

interface FolderMeta {
  path: string[];
  label: string;
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
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [selectedTypeId, setSelectedTypeId] = useState<string | null>(null);
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [selectedFileId, setSelectedFileId] = useState<string | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>(sampleData.toasts);

  const repoTypeMap = useMemo(() => new Map(sampleData.repository_types.map((type) => [type.type, type])), []);
  const selectedType = selectedTypeId ? repoTypeMap.get(selectedTypeId) ?? null : null;

  const builtRepo = useMemo(() => (selectedType ? buildRepoData(selectedType) : null), [selectedType]);

  const folderFiles = selectedFolderId && builtRepo ? builtRepo.filesByFolderId[selectedFolderId] ?? [] : [];
  const selectedFile = selectedFileId && builtRepo ? builtRepo.fileMeta[selectedFileId] ?? null : null;

  const repoTypes: RepoType[] = sampleData.repository_types.map((type) => ({
    id: type.type,
    label: type.label,
    description: type.description,
    onSelect: () => {
      setSelectedTypeId(type.type);
      setSelectedFolderId(null);
      setSelectedFileId(null);
    },
  }));

  const panels: PanelDescriptor[] = [
    {
      id: 'repository-types',
      title: 'Repository Types',
      content: <RepositoryTypesPanel repoTypes={repoTypes} selectedId={selectedTypeId} />,
    },
  ];

  if (selectedType && builtRepo) {
    panels.push({
      id: `folders-${selectedType.type}`,
      title: `${selectedType.label} Folders`,
      content: (
        <FolderBrowserPanel
          folders={builtRepo.folders}
          selectedFolderId={selectedFolderId}
          onFolderSelect={(node) => {
            setSelectedFolderId(node.id);
            setSelectedFileId(null);
          }}
        />
      ),
    });

    panels.push({
      id: `files-${selectedType.type}`,
      title: `${selectedType.label} Files`,
      content: (
        <FileListPanel
          files={folderFiles}
          selectedFileId={selectedFileId}
          onFileSelect={(row) => setSelectedFileId(row.id)}
          emptyFilesMessage={
            selectedFolderId ? 'No files in this folder.' : 'Select a folder on the left to view files.'
          }
        />
      ),
    });
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
    });
  }

  const breadcrumbs: BreadcrumbItem[] = [
    {
      id: 'crumb-types',
      label: 'Repository Types',
      onSelect: () => {
        setSelectedTypeId(null);
        setSelectedFolderId(null);
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
        setSelectedFolderId(null);
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
          onDensityChange={setDensity}
          onThemeChange={setThemeMode}
          onClose={() => setSettingsOpen(false)}
        />
        <ToastQueue toasts={toasts} onDismiss={(id) => setToasts((current) => current.filter((toast) => toast.id !== id))} />
      </Box>
    </ThemeProvider>
  );
}

function buildRepoData(repoType: SampleRepoType): BuiltRepoData {
  const folders: FolderNode[] = [];
  const filesByFolderId: Record<string, FileRow[]> = {};
  const fileMeta: Record<string, FileDetail> = {};
  const folderMeta: Record<string, FolderMeta> = {};

  for (const repo of repoType.repos) {
    const repoNode: FolderNode = {
      id: `repo:${repoType.type}:${repo.name}`,
      name: repo.display_name,
      children: [],
      meta: { path: [repo.display_name] },
    };
    folderMeta[repoNode.id] = { path: [repo.display_name], label: repo.display_name };

    for (const host of repo.hosts) {
      const hostNode: FolderNode = {
        id: `${repoNode.id}:host:${host.host}`,
        name: host.host,
        children: [],
        meta: { path: [repo.display_name, host.host] },
      };
      folderMeta[hostNode.id] = { path: [repo.display_name, host.host], label: host.host };

      for (const group of host.groups) {
        const groupNode: FolderNode = {
          id: `${hostNode.id}:group:${group.group}`,
          name: group.group,
          children: [],
          meta: { path: [repo.display_name, host.host, group.group] },
        };
        folderMeta[groupNode.id] = { path: [repo.display_name, host.host, group.group], label: group.group };

        for (const nameEntry of group.names) {
          const nameNodeId = `${groupNode.id}:name:${nameEntry.name}`;
          const nameNode: FolderNode = {
            id: nameNodeId,
            name: nameEntry.name,
            children: [],
            meta: { path: [repo.display_name, host.host, group.group, nameEntry.name] },
          };
          folderMeta[nameNode.id] = {
            path: [repo.display_name, host.host, group.group, nameEntry.name],
            label: nameEntry.name,
          };

          filesByFolderId[nameNode.id] = nameEntry.files.map((file) => transformFile(repoType, repo, nameEntry, file));
          for (const file of nameEntry.files) {
            const detail = buildFileDetail(repoType, repo, nameEntry, file);
            fileMeta[detail.id] = detail;
          }

          groupNode.children!.push(nameNode);
        }

        hostNode.children!.push(groupNode);
      }

      repoNode.children!.push(hostNode);
    }

    folders.push(repoNode);
  }

  return { folders, filesByFolderId, fileMeta, folderMeta };
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
