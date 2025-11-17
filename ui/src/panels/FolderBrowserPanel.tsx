import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import type { FolderNode } from '../components';
import { FolderTreeView } from '../components';

export interface FolderBrowserPanelProps {
  folders: FolderNode[];
  selectedFolderId?: string | null;
  treeLoading?: boolean;
  treeError?: string;
  onFolderSelect?: (node: FolderNode) => void;
  emptyTreeMessage?: string;
}

export function FolderBrowserPanel({
  folders,
  selectedFolderId,
  treeLoading,
  treeError,
  onFolderSelect,
  emptyTreeMessage = 'No repositories configured for this type.',
}: FolderBrowserPanelProps) {
  const wrapperStyle = {
    display: 'inline-flex',
    flexDirection: 'column',
    backgroundColor: '#e3f2fd',
    minHeight: 'max-content',
    minWidth: 'max-content',
  } as const;

  return (
    <Box sx={wrapperStyle}>
      <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', pr: 1 }}>
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
          <FolderTreeView nodes={folders} selectedId={selectedFolderId} onSelect={onFolderSelect} />
        )}
      </Box>
    </Box>
  );
}
