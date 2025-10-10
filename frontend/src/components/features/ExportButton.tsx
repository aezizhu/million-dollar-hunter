'use client';

import { useState } from 'react';
import {
  Button,
  Menu,
  MenuItem,
  CircularProgress,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import { walletApi } from '@/lib/api';
import { downloadFile } from '@/lib/utils';
import type { Chain } from '@/types';

interface ExportButtonProps {
  address: string;
  chain: Chain;
  defaultFormat?: 'csv' | 'json';
  onError?: (err: Error) => void;
}

export function ExportButton({
  address,
  chain,
  onError,
}: ExportButtonProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [loading, setLoading] = useState(false);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleExport = async (format: 'csv' | 'json') => {
    setLoading(true);
    handleClose();

    try {
      const blob = await walletApi.exportWallet(address, chain, format);
      const filename = `wallet-${address}-${Date.now()}.${format}`;
      downloadFile(blob, filename);
    } catch (error) {
      console.error('Export failed:', error);
      if (onError && error instanceof Error) {
        onError(error);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        variant="outlined"
        startIcon={loading ? <CircularProgress size={20} /> : <DownloadIcon />}
        onClick={handleClick}
        disabled={loading}
      >
        Export
      </Button>
      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
      >
        <MenuItem onClick={() => handleExport('json')}>Export as JSON</MenuItem>
        <MenuItem onClick={() => handleExport('csv')}>Export as CSV</MenuItem>
      </Menu>
    </>
  );
}
