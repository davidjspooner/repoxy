import { useMemo, useState } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import { ListPanelToolbar } from './ListPanelToolbar';

export interface ListPanelItem {
  id: string;
  title: string;
  detail?: string;
}

export interface ListPanelProps {
  items: ListPanelItem[];
  selectedId?: string | null;
  onSelect?: (item: ListPanelItem) => void;
  initialMode?: 'list' | 'tiles';
  emptyMessage?: string;
}

export function ListPanel({
  items,
  selectedId,
  onSelect,
  initialMode = 'list',
  emptyMessage = 'No entries to display.',
}: ListPanelProps) {
  const [mode, setMode] = useState<'list' | 'tiles'>(initialMode);
  const [filterText, setFilterText] = useState('');

  const renderedItems = useMemo(() => {
    const base = items ?? [];
    if (!filterText.trim()) {
      return base;
    }
    const lower = filterText.toLowerCase();
    return base.filter((item) => {
      return (
        item.title.toLowerCase().includes(lower) ||
        (item.detail ? item.detail.toLowerCase().includes(lower) : false)
      );
    });
  }, [items, filterText]);

  return (
    <Box display="flex" flexDirection="column" gap={2}>
      <Box display="flex" justifyContent="flex-start">
        <ListPanelToolbar
          mode={mode}
          onModeChange={setMode}
          filterText={filterText}
          onFilterChange={setFilterText}
        />
      </Box>
      {!renderedItems.length ? (
        <Box textAlign="center" py={4}>
          <Typography variant="body1" color="text.secondary">
            {emptyMessage}
          </Typography>
        </Box>
      ) : mode === 'list' ? (
        <Stack component="ul" spacing={1} sx={{ listStyle: 'none', p: 0, m: 0 }}>
          {renderedItems.map((item) => (
            <Paper
              key={item.id}
              component="li"
              variant={selectedId === item.id ? 'elevation' : 'outlined'}
              sx={{
                p: 2,
                cursor: 'pointer',
                borderColor: selectedId === item.id ? 'primary.main' : undefined,
              }}
              onClick={() => onSelect?.(item)}
            >
              <Typography fontWeight={700}>{item.title}</Typography>
              {item.detail ? (
                <Typography fontStyle="italic" color="text.secondary">
                  {item.detail}
                </Typography>
              ) : null}
            </Paper>
          ))}
        </Stack>
      ) : (
        <Stack direction="row" flexWrap="wrap" spacing={2} useFlexGap>
          {renderedItems.map((item) => (
            <Paper
              key={item.id}
              sx={{
                p: 2,
                width: 200,
                cursor: 'pointer',
                border: selectedId === item.id ? '2px solid' : '1px solid rgba(0,0,0,0.12)',
                borderColor: selectedId === item.id ? 'primary.main' : undefined,
              }}
              onClick={() => onSelect?.(item)}
            >
              <Typography fontWeight={700}>{item.title}</Typography>
              {item.detail ? (
                <Typography fontStyle="italic" color="text.secondary">
                  {item.detail}
                </Typography>
              ) : null}
            </Paper>
          ))}
        </Stack>
      )}
    </Box>
  );
}
