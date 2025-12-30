'use client';

import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Paper,
  Typography,
  Box,
  Chip,
  Link as MuiLink,
} from '@mui/material';
import { formatCurrency, formatDate, formatAddress, getExplorerTxUrl } from '@/lib/utils';
import type { Transaction, Chain } from '@/types';

interface TransactionHistoryProps {
  items: Transaction[];
  page: number;
  pageSize: number;
  total: number;
  chain: Chain;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

const typeColors: Record<Transaction['type'], 'success' | 'error' | 'info' | 'warning' | 'default'> = {
  RECEIVE: 'success',
  SEND: 'error',
  SWAP: 'info',
  APPROVE: 'warning',
  MINT: 'success',
  BURN: 'error',
};

export function TransactionHistory({
  items,
  page,
  pageSize,
  total,
  chain,
  onPageChange,
  onPageSizeChange,
}: TransactionHistoryProps) {
  const handleChangePage = (_event: unknown, newPage: number) => {
    onPageChange(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    onPageSizeChange(parseInt(event.target.value, 10));
    onPageChange(0);
  };

  if (items.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', py: 4 }}>
        <Typography variant="body1" color="text.secondary">
          No transactions found
        </Typography>
      </Box>
    );
  }

  return (
    <Paper sx={{ width: '100%', overflow: 'hidden' }}>
      <TableContainer>
        <Table stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell>Time</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>From</TableCell>
              <TableCell>To</TableCell>
              <TableCell>Asset</TableCell>
              <TableCell align="right">Amount</TableCell>
              <TableCell align="right">Value (USD)</TableCell>
              <TableCell>Tx Hash</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((tx, index) => (
              <TableRow key={`${tx.txHash}-${index}`} hover>
                <TableCell>
                  <Typography variant="body2" noWrap>
                    {formatDate(tx.timestamp)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={tx.type}
                    size="small"
                    color={typeColors[tx.type]}
                    variant="outlined"
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" fontFamily="monospace">
                    {formatAddress(tx.from)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" fontFamily="monospace">
                    {formatAddress(tx.to)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" fontWeight="medium">
                    {tx.symbol}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  {tx.amount.toFixed(4)}
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" fontWeight="medium">
                    {formatCurrency(tx.usdValue)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <MuiLink
                    href={getExplorerTxUrl(tx.txHash, chain)}
                    target="_blank"
                    rel="noopener noreferrer"
                    sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}
                  >
                    {formatAddress(tx.txHash)}
                  </MuiLink>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[25, 50, 100]}
        component="div"
        count={total}
        rowsPerPage={pageSize}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Paper>
  );
}
