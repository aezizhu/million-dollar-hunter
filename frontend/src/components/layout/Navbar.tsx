'use client';

import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  IconButton,
} from '@mui/material';
import Link from 'next/link';
import LogoutIcon from '@mui/icons-material/Logout';
import DashboardIcon from '@mui/icons-material/Dashboard';
import { useAuth } from '@/hooks/useAuth';

export function Navbar() {
  const { isAuthenticated, logout } = useAuth();

  return (
    <AppBar position="static" elevation={0} sx={{ borderBottom: '1px solid', borderColor: 'divider' }}>
      <Toolbar>
        <Typography
          variant="h6"
          component={Link}
          href="/"
          sx={{
            flexGrow: 1,
            textDecoration: 'none',
            color: 'inherit',
            fontWeight: 700,
          }}
        >
          Million Hunter
        </Typography>

        {isAuthenticated && (
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
            <Button
              component={Link}
              href="/dashboard"
              startIcon={<DashboardIcon />}
              color="inherit"
            >
              Dashboard
            </Button>
            <IconButton onClick={logout} color="inherit" title="Logout">
              <LogoutIcon />
            </IconButton>
          </Box>
        )}
      </Toolbar>
    </AppBar>
  );
}
