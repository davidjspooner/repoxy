import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import type { FileRow } from '../components';
import { FileListTable } from '../components';

export interface FileListPanelProps {
  files: FileRow[];
  selectedFileId?: string | null;
  filesLoading?: boolean;
  filesError?: string;
  onFileSelect?: (row: FileRow) => void;
  emptyFilesMessage?: string;
}

export function FileListPanel({
  files,
  selectedFileId,
  filesLoading,
  filesError,
  onFileSelect,
  emptyFilesMessage = 'No files in this folder.',
}: FileListPanelProps) {
  return (
    <Box sx={{ display: 'inline-flex', flexDirection: 'column', minWidth: 'max-content' }}>
      <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
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
          <FileListTable rows={files} onSelect={onFileSelect} selectedFileId={selectedFileId} />
        )}
      </Box>
    </Box>
  );
}
