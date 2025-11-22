import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import type { RepoType, ListPanelItem } from '../components';
import { ListPanel } from '../components';

export interface RepoInstancesPanelProps {
  repos: RepoType[];
  selectedId?: string | null;
  emptyMessage?: string;
}

export function RepoInstancesPanel({ repos, selectedId, emptyMessage = 'No matches.' }: RepoInstancesPanelProps) {
  if (!repos.length) {
    return (
      <Box textAlign="center">
        <Typography variant="body1" color="text.secondary">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  const items: ListPanelItem[] = repos.map((repo) => ({
    id: repo.id,
    title: repo.label,
    detail: repo.description,
  }));
  const repoById = new Map(repos.map((repo) => [repo.id, repo]));

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
