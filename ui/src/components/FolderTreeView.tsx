import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Collapse from '@mui/material/Collapse';
import { ExpandLess, ExpandMore } from '@mui/icons-material';
import { useState } from 'react';
import type { FolderNode } from './types';

export interface FolderTreeViewProps {
  nodes: FolderNode[];
  onSelect?: (node: FolderNode) => void;
}

export function FolderTreeView({ nodes, onSelect }: FolderTreeViewProps) {
  return (
    <List dense disablePadding>
      {nodes.map((node) => (
        <FolderTreeNode key={node.id} node={node} onSelect={onSelect} depth={0} />
      ))}
    </List>
  );
}

interface FolderTreeNodeProps {
  node: FolderNode;
  depth: number;
  onSelect?: (node: FolderNode) => void;
}

function FolderTreeNode({ node, depth, onSelect }: FolderTreeNodeProps) {
  const [open, setOpen] = useState(true);
  const hasChildren = Boolean(node.children?.length);

  return (
    <>
      <ListItemButton
        sx={{ pl: depth * 2 }}
        onClick={() => {
          if (hasChildren) {
            setOpen((prev) => !prev);
          }
          onSelect?.(node);
        }}
      >
        {hasChildren ? (open ? <ExpandLess /> : <ExpandMore />) : null}
        <ListItemText primary={node.name} sx={{ ml: 1 }} />
      </ListItemButton>
      {hasChildren ? (
        <Collapse in={open} timeout="auto" unmountOnExit>
          <List disablePadding>
            {node.children!.map((child) => (
              <FolderTreeNode key={child.id} node={child} onSelect={onSelect} depth={depth + 1} />
            ))}
          </List>
        </Collapse>
      ) : null}
    </>
  );
}
