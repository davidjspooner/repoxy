import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import type { RepoItem, ListPanelItem } from '../components';
import { ListPanel } from '../components';

export interface ItemListPanelProps {
  items: RepoItem[];
  selectedItemId?: string | null;
  onItemSelect?: (item: RepoItem) => void;
  emptyMessage?: string;
  loadingMessage?: string;
}

export function ItemListPanel({
  items,
  selectedItemId,
  onItemSelect,
  emptyMessage = 'No matches.',
  loadingMessage,
}: ItemListPanelProps) {
  if (loadingMessage) {
    return <Alert severity="info">{loadingMessage}</Alert>;
  }

  if (!items.length) {
    return (
      <Box textAlign="center">
        <Typography variant="body1" color="text.secondary">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  const panelItems: ListPanelItem[] = items.map((item) => ({
    id: item.id,
    title: item.label,
    detail: item.description ?? item.path.join(' / '),
  }));

  return (
    <ListPanel
      items={panelItems}
      selectedId={selectedItemId ?? undefined}
      onSelect={(selected) => {
        const found = items.find((item) => item.id === selected.id);
        if (found) {
          onItemSelect?.(found);
        }
      }}
      initialMode="list"
      emptyMessage={emptyMessage}
    />
  );
}
