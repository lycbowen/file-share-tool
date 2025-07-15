import { useEffect, useState } from 'react';

import { List, ListItem, ListItemText, ListItemIcon, IconButton, Button, ListItemButton } from '@mui/material';
import { Box, Paper, Typography, Container, Divider, Toolbar, Grid, DialogTitle, Dialog } from '@mui/material';
import { ContentCopy, FolderOpenOutlined, InsertDriveFileOutlined, IosShareOutlined } from '@mui/icons-material';
import { styled } from '@mui/material/styles';
import { QRCodeCanvas } from 'qrcode.react';

import "./App.css"

const DemoPaper = styled(Paper)(({ theme }) => ({
  width: "80%",
  padding: theme.spacing(2),
  borderRadius: 8,
  ...theme.typography.body2,
  textAlign: 'center',
}));

interface File {
  file_name: string;
  file_modtime: string;
  is_dir: boolean;
  file_size: string; // Update this type based on your actual data
  sub_file_num: number;
  sub_dir_num: number;
}

function normalizePath(base: string, append?: string): string {
  const parts = (base + '/' + (append ?? '')).split('/').filter(Boolean);
  return '/' + parts.join('/');
}

function getParentPath(currentPath: string): string {
  const parts = currentPath.split('/').filter(Boolean);
  parts.pop(); // 去掉最后一节
  return '/' + parts.join('/');
}

function getDisplayPath(fullPath: string, basePath: string): string {
  if (!basePath || !fullPath.startsWith(basePath)) return fullPath;
  const relative = fullPath.substring(basePath.length);
  return relative.startsWith("/") ? relative : "/" + relative;
}

function App() {
  const [files, setFiles] = useState<File[]>([]);
  const [basePath, setBasePath] = useState<string>("");
  const [path, setPath] = useState<string>();
  const [dialogState, setDialogState] = useState<boolean>(false);
  const [shareUrl, setShareUrl] = useState<string>('');
  const [localIp, setLocalIp] = useState("");
  const req = function (path?: string) {
    const query = path ? `?path=${encodeURIComponent(path)}` : '';
    fetch(`/api/files${query}`)
      .then(response => response.json())
      .then(data => {
        if (data.files) {
          setFiles(data.files);
          setLocalIp(data.local_ip);
          setBasePath((prev) => prev || data.path);
          setPath(data.path); // 以后端返回为准
        } else {
          console.warn(data.message);
        }
      })
      .catch(error => console.error('Error:', error));
  };

  useEffect(() => {
    req();
  }, []);

  const clickBtn = function (f: string, isDir: boolean) {
    const fullPath = normalizePath(path ?? '/', f);

    if (!isDir) {
      const fileUrl = `/api/download?fname=${fullPath}`;
      const link = document.createElement('a');
      link.href = fileUrl;
      link.download = f;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      return;
    }
    req(fullPath);
  };

  const share = function (fname: string) {
    if (path !== '/') {
      setShareUrl(`http://${localIp}/api/download?fname=${path}/${fname}`)
    } else {
      setShareUrl(`http://${localIp}/api/download?fname=/${fname}`)
    }
    setDialogState(true);
  }

  const clickBackBtn = function () {
    if (!path || path === "/") return;

    const parent = getParentPath(path);
    req(parent);
  };

  return (
    <Container maxWidth="md">
      <Box style={{
        display: 'flex',
        minWidth: "650px",
        minHeight: "100vh",
        flexDirection: "column",
        alignItems: "center",
      }}>
        <Typography variant="h4" gutterBottom style={{ margin: "30px 80px", alignSelf: "flex-start" }}>
          Index of <code>{getDisplayPath(path ?? "", basePath)}</code>
        </Typography>
        <DemoPaper elevation={6} square={false}>
          <List dense={false} >
            <ListItem key={-1} secondaryAction={<IconButton edge="end" aria-label="more"></IconButton>} disablePadding>
              <ListItemButton onClick={clickBackBtn}>
                <ListItemIcon>
                  <FolderOpenOutlined />
                </ListItemIcon>
                <Box sx={{ flexGrow: 1 }}>
                  <Grid container spacing={2}>
                    <Grid size={{ xs: 6 }}>
                      <ListItemText primary={'../'} />
                    </Grid>
                    <Grid size={{ xs: 4 }}>
                    </Grid>
                    <Grid size={{ xs: 2 }}>
                    </Grid>
                  </Grid>
                </Box>

              </ListItemButton>
            </ListItem>
            {files.map((item, index) => (
              [<Divider />,
              <ListItem key={index} secondaryAction={
                <IconButton onClick={() => { share(item.file_name) }} edge="end" aria-label="more">
                  <IosShareOutlined style={{ display: item.is_dir ? 'none' : 'block' }} />
                </IconButton>} disablePadding>
                <ListItemButton onClick={() => { clickBtn(item.file_name, item.is_dir) }}>
                  <ListItemIcon>
                    {item.is_dir ? <FolderOpenOutlined /> : <InsertDriveFileOutlined />}
                  </ListItemIcon>
                  <Box sx={{ flexGrow: 1 }}>
                    <Grid container spacing={2}>
                      <Grid size={{ xs: 6 }}>
                        <ListItemText primary={item.file_name} />
                      </Grid>
                      <Grid size={{ xs: 4 }}>
                        <ListItemText secondary={item.file_modtime} />
                      </Grid>
                      <Grid size={{ xs: 2 }}>
                        <ListItemText secondary={item.is_dir ? `${item.sub_dir_num} dirs ${item.sub_file_num} files` : item.file_size} />
                      </Grid>
                    </Grid>
                  </Box>

                </ListItemButton>
              </ListItem>]
            ))}
          </List>
        </DemoPaper>
        <Toolbar style={{ flexShrink: 0 }}>
          <Typography variant="body1" color="inherit">
            © File Share Tool. By <a href={`mailto:ischenbowen@outlook.com`} onClick={() => { window.location.href = `mailto:ischenbowen@outlook.com` }}>cbowen</a>
          </Typography>
        </Toolbar>
        <Dialog open={dialogState} onClose={() => { setDialogState(false) }}>
          <DialogTitle>
            <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
              <QRCodeCanvas value={shareUrl} />
              <Button
                variant="text"
                startIcon={<ContentCopy />}
                onClick={() => navigator.clipboard.writeText(shareUrl)}
              >
                复制链接
              </Button>
            </div>
          </DialogTitle>
        </Dialog>
      </Box>
    </Container>

  );
}

export default App;
