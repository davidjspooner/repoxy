import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Unstable_Grid2';
import type { FolderNode, FileRow } from '../components';
import { FolderTreeView, FileListTable } from '../components';

export interface RepositoryBrowserPanelProps {
  folders: FolderNode[];
  files: FileRow[];
  treeLoading?: boolean;
  filesLoading?: boolean;
  treeError?: string;
  filesError?: string;
  onFolderSelect?: (node: FolderNode) => void;
  onFileSelect?: (row: FileRow) => void;
  emptyTreeMessage?: string;
  emptyFilesMessage?: string;
}

export function RepositoryBrowserPanel({
  folders,
  files,
  treeLoading,
  filesLoading,
  treeError,
  filesError,
  onFolderSelect,
  onFileSelect,
  emptyTreeMessage = 'No repositories configured for this type.',
  emptyFilesMessage = 'No files in this folder.',
}: RepositoryBrowserPanelProps) {
  return (
    <Grid container spacing={2} height="100%">
      <Grid xs={12} md={5} lg={4} xl={3}>
        <Box height="100%" borderRight={{ md: '1px solid #1f2a36' }} pr={{ md: 2 }}>
          {treeLoading ? (
            <Box display="flex" justifyContent="center" py={2}>
              <CircularProgress size={24} />
            </Box>
          ) : treeError ? (
            <Alert severity="error">{treeError}</Alert>
          ) : folders.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              {emptyTreeMessage}
            </Typography>
          ) : (
            <FolderTreeView nodes={folders} onSelect={onFolderSelect} />
          )}
        </Box>
      </Grid>
      <Grid xs={12} md={7} lg={8} xl={9}>
        {filesLoading ? (
          <Box display="flex" justifyContent="center" py={2}>
            <CircularProgress size={24} />
          </Box>
        ) : filesError ? (
          <Alert severity="error">{filesError}</Alert>
        ) : files.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {emptyFilesMessage}
          </Typography>
        ) : (
          <FileListTable rows={files} onSelect={onFileSelect} />
        )}
      </Grid>
    </Grid>
  );
}
