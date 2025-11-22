import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import type { RepoItemVersion, ListPanelItem } from '../components';
import { ListPanel } from '../components';

export interface VersionListPanelProps {
  versions: RepoItemVersion[];
  selectedVersionId?: string | null;
  onVersionSelect?: (version: RepoItemVersion) => void;
  emptyMessage?: string;
}

export function VersionListPanel({
  versions,
  selectedVersionId,
  onVersionSelect,
  emptyMessage = 'No matches.',
}: VersionListPanelProps) {
  if (!versions.length) {
    return (
      <Box textAlign="center">
        <Typography variant="body1" color="text.secondary">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  const panelItems: ListPanelItem[] = versions.map((version) => ({
    id: version.id,
    title: version.label,
    detail: version.description,
  }));

  return (
    <ListPanel
      items={panelItems}
      selectedId={selectedVersionId ?? undefined}
      onSelect={(selection) => {
        const found = versions.find((version) => version.id === selection.id);
        if (found) {
          onVersionSelect?.(found);
        }
      }}
      initialMode="list"
      emptyMessage={emptyMessage}
    />
  );
}
