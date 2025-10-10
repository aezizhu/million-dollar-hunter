'use client';

import { Card, CardContent, Typography, Box, Chip } from '@mui/material';
import Link from 'next/link';
import { formatCurrency, formatPercent, formatAddress } from '@/lib/utils';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import TrendingDownIcon from '@mui/icons-material/TrendingDown';

interface WalletCardProps {
  address: string;
  nickname?: string;
  netWorthUsd: number;
  changePct24h?: number;
  chain: string;
}

export function WalletCard({
  address,
  nickname,
  netWorthUsd,
  changePct24h,
  chain,
}: WalletCardProps) {
  const isPositiveChange = changePct24h !== undefined && changePct24h >= 0;

  return (
    <Link
      href={`/wallets/${address}?chain=${chain}`}
      style={{ textDecoration: 'none' }}
    >
      <Card
        sx={{
          height: '100%',
          cursor: 'pointer',
          transition: 'all 0.2s',
          border: '1px solid',
          borderColor: 'divider',
          '&:hover': {
            borderColor: 'primary.main',
            transform: 'translateY(-2px)',
          },
        }}
      >
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6" component="div" noWrap>
              {nickname || formatAddress(address)}
            </Typography>
            <Chip
              label={chain.toUpperCase()}
              size="small"
              color="primary"
              variant="outlined"
            />
          </Box>

          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            {formatAddress(address)}
          </Typography>

          <Typography variant="h4" component="div" sx={{ mb: 1 }}>
            {formatCurrency(netWorthUsd)}
          </Typography>

          {changePct24h !== undefined && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
              {isPositiveChange ? (
                <TrendingUpIcon color="success" fontSize="small" />
              ) : (
                <TrendingDownIcon color="error" fontSize="small" />
              )}
              <Typography
                variant="body2"
                sx={{
                  color: isPositiveChange ? 'success.main' : 'error.main',
                }}
              >
                {formatPercent(changePct24h)} (24h)
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
