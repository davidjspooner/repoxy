import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import IconButton from '@mui/material/IconButton';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import { useState } from 'react';
import type { FileRow } from './types';

export interface FileListTableProps {
  rows: FileRow[];
  selectedFileId?: string | null;
  onSelect?: (row: FileRow) => void;
}

type SortKey = 'name' | 'modified' | 'sizeBytes';

export function FileListTable({ rows, onSelect, selectedFileId }: FileListTableProps) {
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [direction, setDirection] = useState<'asc' | 'desc'>('asc');

  const sortedRows = [...rows].sort((a, b) => {
    const valueA = a[sortKey];
    const valueB = b[sortKey];
    if (valueA < valueB) return direction === 'asc' ? -1 : 1;
    if (valueA > valueB) return direction === 'asc' ? 1 : -1;
    return 0;
  });

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setDirection('asc');
    }
  };

  const sortIcon = (key: SortKey) => {
    if (sortKey !== key) return null;
    return direction === 'asc' ? <ArrowUpwardIcon fontSize="inherit" /> : <ArrowDownwardIcon fontSize="inherit" />;
  };

  return (
    <TableContainer component={Paper} elevation={0} sx={{ backgroundColor: 'transparent' }}>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>
              <IconButton size="small" onClick={() => toggleSort('name')}>
                Name {sortIcon('name')}
              </IconButton>
            </TableCell>
            <TableCell>
              <IconButton size="small" onClick={() => toggleSort('modified')}>
                Modified {sortIcon('modified')}
              </IconButton>
            </TableCell>
            <TableCell align="right">
              <IconButton size="small" onClick={() => toggleSort('sizeBytes')}>
                Size {sortIcon('sizeBytes')}
              </IconButton>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedRows.map((row) => (
            <TableRow
              key={row.id}
              hover
              selected={row.id === selectedFileId}
              sx={{ cursor: 'pointer' }}
              onClick={() => onSelect?.(row)}
            >
              <TableCell>{row.name}</TableCell>
              <TableCell>{row.modified}</TableCell>
              <TableCell align="right">{formatBytes(row.sizeBytes)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

function formatBytes(size: number) {
  if (size < 1024) return `${size} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let value = size / 1024;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`;
}
