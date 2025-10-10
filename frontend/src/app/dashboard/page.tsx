'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Container,
  Box,
  Typography,
  Button,
  Grid,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portfolioApi } from '@/lib/api';
import { WalletCard } from '@/components/features/WalletCard';
import { useAuth } from '@/hooks/useAuth';
import type { Chain } from '@/types';
import { useEffect } from 'react';

const chains: { value: Chain; label: string }[] = [
  { value: 'ethereum', label: 'Ethereum' },
  { value: 'bsc', label: 'BSC' },
  { value: 'polygon', label: 'Polygon' },
  { value: 'arbitrum', label: 'Arbitrum' },
  { value: 'optimism', label: 'Optimism' },
  { value: 'solana', label: 'Solana' },
];

export default function DashboardPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const queryClient = useQueryClient();
  const [openDialog, setOpenDialog] = useState(false);
  const [address, setAddress] = useState('');
  const [chain, setChain] = useState<Chain>('ethereum');
  const [nickname, setNickname] = useState('');
  const [mutationError, setMutationError] = useState('');

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  const { data, isLoading, error } = useQuery({
    queryKey: ['portfolios'],
    queryFn: () => portfolioApi.list(1, 20),
    enabled: isAuthenticated,
  });

  const addWalletMutation = useMutation({
    mutationFn: portfolioApi.addWallet,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portfolios'] });
      setOpenDialog(false);
      setAddress('');
      setNickname('');
      setChain('ethereum');
      setMutationError('');
    },
    onError: () => {
      setMutationError('Failed to add wallet. Please try again.');
    },
  });

  const validateAddress = (addr: string, chain: Chain): boolean => {
    if (chain === 'solana') {
      return /^[1-9A-HJ-NP-Za-km-z]{32,44}$/.test(addr);
    }
    return /^0x[a-fA-F0-9]{40}$/.test(addr);
  };

  const handleAddWallet = () => {
    if (!address) {
      setMutationError('Please enter a wallet address');
      return;
    }
    
    if (!validateAddress(address, chain)) {
      setMutationError(`Invalid ${chain} address format`);
      return;
    }
    
    setMutationError('');
    addWalletMutation.mutate({ address, chain, nickname: nickname || undefined });
  };

  if (authLoading) {
    return (
      <Container>
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h4" component="h1">
          Portfolio Dashboard
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpenDialog(true)}
        >
          Add Wallet
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 4 }}>
          Failed to load portfolios
        </Alert>
      )}

      {isLoading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      )}

      {data && data.items.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No wallets tracked yet
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Add a wallet to start tracking your portfolio
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
          >
            Add Your First Wallet
          </Button>
        </Box>
      )}

      {data && data.items.length > 0 && (
        <Grid container spacing={3}>
          {data.items.map((portfolio) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={portfolio.id}>
              <WalletCard
                address={portfolio.address}
                nickname={portfolio.nickname}
                netWorthUsd={portfolio.netWorthUsd}
                chain={portfolio.chain}
              />
            </Grid>
          ))}
        </Grid>
      )}

      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Wallet to Track</DialogTitle>
        <DialogContent>
          {mutationError && (
            <Alert severity="error" sx={{ mt: 2, mb: 2 }}>
              {mutationError}
            </Alert>
          )}
          <TextField
            label="Wallet Address"
            fullWidth
            required
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            sx={{ mt: 2, mb: 2 }}
            placeholder="0x... or Solana address"
          />
          <TextField
            label="Chain"
            fullWidth
            required
            select
            value={chain}
            onChange={(e) => setChain(e.target.value as Chain)}
            sx={{ mb: 2 }}
          >
            {chains.map((option) => (
              <MenuItem key={option.value} value={option.value}>
                {option.label}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            label="Nickname (Optional)"
            fullWidth
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            placeholder="My Main Wallet"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button
            onClick={handleAddWallet}
            variant="contained"
            disabled={!address || addWalletMutation.isPending}
          >
            {addWalletMutation.isPending ? 'Adding...' : 'Add Wallet'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
}
