'use client';

import { useState, use } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import dynamic from 'next/dynamic';
import {
  Container,
  Box,
  Typography,
  Paper,
  Grid,
  CircularProgress,
  Alert,
  Tabs,
  Tab,
  Chip,
} from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import { walletApi } from '@/lib/api';
import { AssetHoldings } from '@/components/features/AssetHoldings';
import { TransactionHistory } from '@/components/features/TransactionHistory';
import { ExportButton } from '@/components/features/ExportButton';
import { formatCurrency, formatAddress } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';
import type { Chain } from '@/types';
import { useEffect } from 'react';

const FinancialChart = dynamic(
  () => import('@/components/charts/FinancialChart').then((mod) => ({ default: mod.FinancialChart })),
  { ssr: false, loading: () => <CircularProgress /> }
);

interface PageProps {
  params: Promise<{ address: string }>;
}

export default function WalletPage({ params }: PageProps) {
  const resolvedParams = use(params);
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [activeTab, setActiveTab] = useState(0);
  const [assetPage, setAssetPage] = useState(0);
  const [assetPageSize, setAssetPageSize] = useState(10);
  const [txPage, setTxPage] = useState(1);
  const [txPageSize, setTxPageSize] = useState(50);

  const chain = (searchParams.get('chain') as Chain) || 'ethereum';

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  const { data: walletData, isLoading: walletLoading, error: walletError } = useQuery({
    queryKey: ['wallet', resolvedParams.address, chain],
    queryFn: () => walletApi.getWallet(resolvedParams.address, chain),
    enabled: isAuthenticated,
  });

  const { data: txData, isLoading: txLoading } = useQuery({
    queryKey: ['transactions', resolvedParams.address, txPage, txPageSize],
    queryFn: () => walletApi.getTransactions(resolvedParams.address, txPage, txPageSize),
    enabled: isAuthenticated && activeTab === 1,
  });

  if (authLoading || walletLoading) {
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

  if (walletError) {
    return (
      <Container sx={{ py: 4 }}>
        <Alert severity="error">Failed to load wallet data</Alert>
      </Container>
    );
  }

  if (!walletData) {
    return null;
  }

  const displayedAssets = walletData.assets.slice(
    assetPage * assetPageSize,
    (assetPage + 1) * assetPageSize
  );

  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography variant="h4" component="h1">
              {walletData.nickname || 'Wallet Details'}
            </Typography>
            <Chip label={chain.toUpperCase()} color="primary" />
          </Box>
          <ExportButton address={resolvedParams.address} chain={chain} />
        </Box>
        <Typography variant="body2" color="text.secondary" fontFamily="monospace">
          {formatAddress(walletData.address, 10, 8)}
        </Typography>
      </Box>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Net Worth
            </Typography>
            <Typography variant="h3" component="div">
              {formatCurrency(walletData.netWorthUsd || 0)}
            </Typography>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Total Assets
            </Typography>
            <Typography variant="h3" component="div">
              {walletData.assets.length}
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      {walletData.history && walletData.history.length > 0 && (
        <Paper sx={{ p: 3, mb: 4 }}>
          <FinancialChart
            data={walletData.history}
            title="Portfolio Value"
            height={400}
          />
        </Paper>
      )}

      <Paper sx={{ mb: 4 }}>
        <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)}>
          <Tab label="Assets" />
          <Tab label="Transactions" />
        </Tabs>

        <Box sx={{ p: 3 }}>
          {activeTab === 0 && (
            <AssetHoldings
              items={displayedAssets}
              page={assetPage}
              pageSize={assetPageSize}
              onPageChange={setAssetPage}
              onPageSizeChange={setAssetPageSize}
            />
          )}

          {activeTab === 1 && (
            <>
              {txLoading && (
                <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                  <CircularProgress />
                </Box>
              )}
              {txData && (
                <TransactionHistory
                  items={txData.items}
                  page={txPage - 1}
                  pageSize={txPageSize}
                  total={txData.total}
                  chain={chain}
                  onPageChange={(p) => setTxPage(p + 1)}
                  onPageSizeChange={setTxPageSize}
                />
              )}
            </>
          )}
        </Box>
      </Paper>
    </Container>
  );
}
