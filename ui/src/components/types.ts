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
}

export interface FolderNode {
  id: string;
  name: string;
  children?: FolderNode[];
}

export interface FileRow {
  id: string;
  name: string;
  modified: string;
  sizeBytes: number;
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
