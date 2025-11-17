import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Skeleton from '@mui/material/Skeleton';
import Typography from '@mui/material/Typography';
import type { RepoType } from '../components';
import { RepoTypeTile, TileGrid } from '../components';

export interface RepositoryTypesPanelProps {
  repoTypes: RepoType[];
  selectedId?: string | null;
  loading?: boolean;
  error?: string;
  emptyMessage?: string;
}

export function RepositoryTypesPanel({
  repoTypes,
  selectedId,
  loading,
  error,
  emptyMessage = 'No cached repositories available.',
}: RepositoryTypesPanelProps) {
  const wrapperStyle = {
    display: 'inline-flex',
    flexDirection: 'column',
    minHeight: 'max-content',
    minWidth: 'max-content',
    backgroundColor: '#e3f2fd',
  } as const;

  if (loading) {
    return (
      <Box sx={wrapperStyle}>
        <TileGrid>
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} variant="rectangular" height={120} />
          ))}
        </TileGrid>
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={wrapperStyle}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  if (!repoTypes.length) {
    return (
      <Box sx={{ ...wrapperStyle, textAlign: 'center' }}>
        <Typography variant="body1" color="text.secondary">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={wrapperStyle}>
      <TileGrid>
        {repoTypes.map((repo) => (
          <RepoTypeTile key={repo.id} repoType={repo} selected={selectedId === repo.id} />
        ))}
      </TileGrid>
    </Box>
  );
}
