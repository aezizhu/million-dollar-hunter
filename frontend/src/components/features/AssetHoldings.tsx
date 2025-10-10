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
} from '@mui/material';
import { formatCurrency, formatNumber } from '@/lib/utils';
import type { Asset } from '@/types';

interface AssetHoldingsProps {
  items: Asset[];
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

export function AssetHoldings({
  items,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
}: AssetHoldingsProps) {
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
          No assets found
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
              <TableCell>Asset</TableCell>
              <TableCell>Symbol</TableCell>
              <TableCell align="right">Balance</TableCell>
              <TableCell align="right">Value (USD)</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((asset) => (
              <TableRow key={asset.tokenAddress} hover>
                <TableCell>
                  <Typography variant="body2" noWrap>
                    {asset.name || asset.symbol}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" fontWeight="medium">
                    {asset.symbol}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  {formatNumber(asset.balance, 4)}
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" fontWeight="medium">
                    {formatCurrency(asset.usdValue)}
                  </Typography>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[10, 25, 50]}
        component="div"
        count={items.length}
        rowsPerPage={pageSize}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Paper>
  );
}
