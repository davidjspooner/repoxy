import type React from 'react';

export interface BreadcrumbItem {
  id: string;
  label: string;
  onSelect?: () => void;
  isCurrent?: boolean;
}

export interface PanelDescriptor {
  id: string;
  title: string;
  content: React.ReactNode;
  /**
   * When false, the panel is skipped when selecting the single visible panel on mobile.
   * Defaults to true.
   */
  mobileVisible?: boolean;
}

export interface FileRow {
  id: string;
  name: string;
  modified: string;
  sizeBytes: number;
  path?: string;
}

export interface RepoItem {
  id: string;
  label: string;
  description?: string;
  path: string[];
}

export interface RepoItemVersion {
  id: string;
  itemId: string;
  label: string;
  description?: string;
}

export interface ToastMessage {
  id: string;
  message: string;
  level: 'success' | 'error' | 'info';
}

export interface RepoType {
  id: string;
  label: string;
  description?: string;
  onSelect?: () => void;
}

export interface FolderNode {
  id: string;
  name: string;
  children?: FolderNode[];
  meta?: Record<string, unknown>;
}
