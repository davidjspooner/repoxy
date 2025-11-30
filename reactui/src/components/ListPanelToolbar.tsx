import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import SearchIcon from '@mui/icons-material/Search';
import CloseIcon from '@mui/icons-material/Close';
import type { Dispatch, SetStateAction } from 'react';

export interface ListPanelToolbarProps {
  mode: 'list' | 'tiles';
  onModeChange: (mode: 'list' | 'tiles') => void;
  filterText: string;
  onFilterChange: Dispatch<SetStateAction<string>>;
  filterPlaceholder?: string;
}

export function ListPanelToolbar({
  mode,
  onModeChange,
  filterText,
  onFilterChange,
  filterPlaceholder = 'Filter items…',
}: ListPanelToolbarProps) {
  return (
    <Box display="flex" gap={2} alignItems="center">
      <ToggleButtonGroup
        size="small"
        value={mode}
        exclusive
        onChange={(_, value) => value && onModeChange(value)}
        aria-label="List display mode"
      >
        <ToggleButton value="list" aria-label="List mode">
          List
        </ToggleButton>
        <ToggleButton value="tiles" aria-label="Tile mode">
          Tiles
        </ToggleButton>
      </ToggleButtonGroup>
      <TextField
        variant="outlined"
        size="small"
        placeholder={filterPlaceholder}
        value={filterText}
        onChange={(event) => onFilterChange(event.target.value)}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon fontSize="small" />
            </InputAdornment>
          ),
          endAdornment: filterText ? (
            <InputAdornment position="end">
              <IconButton aria-label="Clear filter" edge="end" size="small" onClick={() => onFilterChange('')}>
                <CloseIcon fontSize="small" />
              </IconButton>
            </InputAdornment>
          ) : undefined,
        }}
        sx={{ minWidth: 220 }}
      />
    </Box>
  );
}
