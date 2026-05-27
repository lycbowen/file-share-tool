import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  AppBar,
  Box,
  Breadcrumbs,
  Button,
  Chip,
  CircularProgress,
  Container,
  Dialog,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  Link,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Paper,
  Snackbar,
  Stack,
  Toolbar,
  Tooltip,
  Typography,
} from '@mui/material';
import {
  ArrowBack,
  ContentCopy,
  FolderOpenOutlined,
  HomeOutlined,
  InsertDriveFileOutlined,
  IosShareOutlined,
  Refresh,
} from '@mui/icons-material';
import { QRCodeCanvas } from 'qrcode.react';

import './App.css';

interface FileItem {
  file_name: string;
  file_modtime: string;
  is_dir: boolean;
  file_size: string;
  sub_file_num: number;
  sub_dir_num: number;
  path: string;
}

interface FilesResponse {
  message?: string;
  local_ip: string;
  public_url: string;
  path: string;
  files: FileItem[];
}

function normalizePath(path: string): string {
  return path.replace(/^\/+/, '').replace(/\/+$/, '');
}

function parentPath(path: string): string {
  const parts = normalizePath(path).split('/').filter(Boolean);
  parts.pop();
  return parts.join('/');
}

function pathParts(path: string): string[] {
  return normalizePath(path).split('/').filter(Boolean);
}

function pathThrough(parts: string[], index: number): string {
  return parts.slice(0, index + 1).join('/');
}

function apiPath(path: string): string {
  const clean = normalizePath(path);
  return clean ? `?path=${encodeURIComponent(clean)}` : '';
}

function App() {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [path, setPath] = useState('');
  const [publicUrl, setPublicUrl] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [shareUrl, setShareUrl] = useState('');
  const [shareName, setShareName] = useState('');
  const [toastOpen, setToastOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState('');

  const crumbs = useMemo(() => pathParts(path), [path]);

  async function loadFiles(nextPath = '') {
    setLoading(true);
    setError('');
    try {
      const response = await fetch(`/api/files${apiPath(nextPath)}`);
      const data = (await response.json()) as FilesResponse;
      if (!response.ok) {
        throw new Error(data.message || '无法读取目录');
      }
      setFiles(data.files ?? []);
      setPath(data.path ?? '');
      setPublicUrl(data.public_url || `http://${data.local_ip}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : '请求失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadFiles();
  }, []);

  function openItem(item: FileItem) {
    if (item.is_dir) {
      void loadFiles(item.path);
      return;
    }
    window.location.href = `/api/download?path=${encodeURIComponent(item.path)}`;
  }

  function openShare(item: FileItem) {
    const base = publicUrl || window.location.origin;
    setShareName(item.file_name);
    setShareUrl(`${base}/api/download?path=${encodeURIComponent(item.path)}`);
  }

  async function copyShareUrl() {
    try {
      await copyText(shareUrl);
      setToastMessage('链接已复制');
    } catch {
      setToastMessage('复制失败，请手动长按或选中链接复制');
    }
    setToastOpen(true);
  }

  return (
    <Box className="app-shell">
      <AppBar position="sticky" color="default" elevation={0} className="topbar">
        <Toolbar className="topbar-inner">
          <Stack spacing={0.25} sx={{ minWidth: 0 }}>
            <Typography variant="h6" noWrap>
              局域网传输助手
            </Typography>
            <Typography variant="caption" color="text.secondary" noWrap>
              {publicUrl || '正在准备局域网访问地址'}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Tooltip title="返回上级">
              <span>
                <IconButton disabled={!path || loading} onClick={() => void loadFiles(parentPath(path))}>
                  <ArrowBack />
                </IconButton>
              </span>
            </Tooltip>
            <Tooltip title="刷新">
              <span>
                <IconButton disabled={loading} onClick={() => void loadFiles(path)}>
                  <Refresh />
                </IconButton>
              </span>
            </Tooltip>
          </Stack>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" className="content">
        <Stack spacing={2.5}>
          <Paper className="path-panel" variant="outlined">
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ justifyContent: 'space-between' }}>
              <Breadcrumbs aria-label="当前路径" className="breadcrumbs">
                <Link component="button" underline="hover" color="inherit" onClick={() => void loadFiles('')}>
                  <Stack direction="row" spacing={0.5} sx={{ alignItems: 'center' }}>
                    <HomeOutlined fontSize="small" />
                    <span>共享根目录</span>
                  </Stack>
                </Link>
                {crumbs.map((part, index) => (
                  <Link
                    component="button"
                    underline="hover"
                    color={index === crumbs.length - 1 ? 'text.primary' : 'inherit'}
                    key={pathThrough(crumbs, index)}
                    onClick={() => void loadFiles(pathThrough(crumbs, index))}
                  >
                    {part}
                  </Link>
                ))}
              </Breadcrumbs>
              <Chip size="small" label={`${files.length} 项`} />
            </Stack>
          </Paper>

          {error && <Alert severity="error">{error}</Alert>}

          <Paper className="file-panel" variant="outlined">
            {loading ? (
              <Stack className="empty-state" sx={{ alignItems: 'center', justifyContent: 'center' }}>
                <CircularProgress size={28} />
                <Typography color="text.secondary">正在读取目录...</Typography>
              </Stack>
            ) : files.length === 0 ? (
              <Stack className="empty-state" sx={{ alignItems: 'center', justifyContent: 'center' }}>
                <FolderOpenOutlined color="disabled" fontSize="large" />
                <Typography color="text.secondary">这个目录是空的</Typography>
              </Stack>
            ) : (
              <List disablePadding>
                {files.map((item, index) => (
                  <Box key={item.path || item.file_name}>
                    {index > 0 && <Divider component="li" />}
                    <ListItem
                      disablePadding
                      secondaryAction={
                        !item.is_dir && (
                          <Tooltip title="分享下载链接">
                            <IconButton edge="end" onClick={() => openShare(item)}>
                              <IosShareOutlined />
                            </IconButton>
                          </Tooltip>
                        )
                      }
                    >
                      <ListItemButton className="file-row" onClick={() => openItem(item)}>
                        <ListItemIcon>{item.is_dir ? <FolderOpenOutlined /> : <InsertDriveFileOutlined />}</ListItemIcon>
                        <ListItemText
                          primary={<Typography noWrap>{item.file_name}</Typography>}
                          secondary={
                            <Typography variant="body2" color="text.secondary" noWrap>
                              {item.file_modtime}
                            </Typography>
                          }
                        />
                        <Typography className="file-meta" color="text.secondary" variant="body2">
                          {item.is_dir ? `${item.sub_dir_num} 个文件夹 · ${item.sub_file_num} 个文件` : item.file_size}
                        </Typography>
                      </ListItemButton>
                    </ListItem>
                  </Box>
                ))}
              </List>
            )}
          </Paper>
        </Stack>
      </Container>

      <Dialog open={Boolean(shareUrl)} onClose={() => setShareUrl('')} fullWidth maxWidth="xs">
        <DialogTitle>分享 {shareName}</DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ alignItems: 'center' }}>
            <QRCodeCanvas value={shareUrl} size={180} />
            <Typography variant="body2" color="text.secondary" className="share-url">
              {shareUrl}
            </Typography>
            <Button variant="contained" startIcon={<ContentCopy />} onClick={() => void copyShareUrl()}>
              复制链接
            </Button>
          </Stack>
        </DialogContent>
      </Dialog>

      <Snackbar open={toastOpen} autoHideDuration={2200} onClose={() => setToastOpen(false)} message={toastMessage} />
    </Box>
  );
}

async function copyText(value: string) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(value);
    return;
  }

  const textarea = document.createElement('textarea');
  textarea.value = value;
  textarea.setAttribute('readonly', '');
  textarea.style.position = 'fixed';
  textarea.style.left = '-9999px';
  textarea.style.top = '0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();

  try {
    const copied = document.execCommand('copy');
    if (!copied) {
      throw new Error('copy command failed');
    }
  } finally {
    document.body.removeChild(textarea);
  }
}

export default App;
