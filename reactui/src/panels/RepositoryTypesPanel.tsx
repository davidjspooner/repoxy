import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Skeleton from '@mui/material/Skeleton';
import Typography from '@mui/material/Typography';
import type { RepoType, ListPanelItem } from '../components';
import { ListPanel, TileGrid } from '../components';

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
  emptyMessage = 'No matches.',
}: RepositoryTypesPanelProps) {
  if (loading) {
    return (
      <TileGrid>
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} variant="rectangular" height={120} />
        ))}
      </TileGrid>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (!repoTypes.length) {
    return (
      <Box textAlign="center">
        <Typography variant="body1" color="text.secondary">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  const items: ListPanelItem[] = repoTypes.map((repo) => ({
    id: repo.id,
    title: repo.label,
    detail: repo.description,
  }));

  const repoById = new Map(repoTypes.map((repo) => [repo.id, repo]));

  return (
    <ListPanel
      items={items}
      selectedId={selectedId ?? undefined}
      onSelect={(item) => repoById.get(item.id)?.onSelect?.()}
      initialMode="tiles"
      emptyMessage={emptyMessage}
    />
  );
}
