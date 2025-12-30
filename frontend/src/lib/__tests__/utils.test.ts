import {
  formatCurrency,
  formatNumber,
  formatPercent,
  formatAddress,
  isValidEthAddress,
  isValidSolanaAddress,
  isValidAddress,
  getExplorerTxUrl,
  getExplorerAddressUrl,
} from '../utils';

describe('utils', () => {
  describe('formatCurrency', () => {
    it('formats currency correctly', () => {
      expect(formatCurrency(1234.56)).toBe('$1,234.56');
      expect(formatCurrency(0)).toBe('$0.00');
      expect(formatCurrency(1234.567, 3)).toBe('$1,234.567');
    });
  });

  describe('formatNumber', () => {
    it('formats numbers correctly', () => {
      expect(formatNumber(1234.56)).toBe('1,234.56');
      expect(formatNumber(1234.567, 3)).toBe('1,234.567');
    });
  });

  describe('formatPercent', () => {
    it('formats percentages correctly', () => {
      expect(formatPercent(5.5)).toBe('+5.50%');
      expect(formatPercent(-3.2)).toBe('-3.20%');
      expect(formatPercent(0)).toBe('+0.00%');
    });
  });

  describe('formatAddress', () => {
    it('formats ethereum address', () => {
      const address = '0x1234567890abcdef1234567890abcdef12345678';
      expect(formatAddress(address)).toBe('0x1234...5678');
    });

    it('formats with custom lengths', () => {
      const address = '0x1234567890abcdef1234567890abcdef12345678';
      expect(formatAddress(address, 4, 4)).toBe('0x12...5678');
    });

    it('returns full address if too short', () => {
      const address = '0x12345';
      expect(formatAddress(address)).toBe('0x12345');
    });
  });

  describe('isValidEthAddress', () => {
    it('validates ethereum addresses', () => {
      expect(isValidEthAddress('0x1234567890abcdef1234567890abcdef12345678')).toBe(true);
      expect(isValidEthAddress('0xInvalidAddress')).toBe(false);
      expect(isValidEthAddress('1234567890abcdef1234567890abcdef12345678')).toBe(false);
    });
  });

  describe('isValidSolanaAddress', () => {
    it('validates solana addresses', () => {
      expect(isValidSolanaAddress('DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK')).toBe(true);
      expect(isValidSolanaAddress('0x1234567890abcdef1234567890abcdef12345678')).toBe(false);
      expect(isValidSolanaAddress('invalid')).toBe(false);
    });
  });

  describe('isValidAddress', () => {
    it('validates both ethereum and solana addresses', () => {
      expect(isValidAddress('0x1234567890abcdef1234567890abcdef12345678')).toBe(true);
      expect(isValidAddress('DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK')).toBe(true);
      expect(isValidAddress('invalid')).toBe(false);
    });
  });

  describe('getExplorerTxUrl', () => {
    it('returns correct explorer URL for ethereum transactions', () => {
      const txHash = '0x123abc';
      expect(getExplorerTxUrl(txHash, 'ethereum')).toBe('https://etherscan.io/tx/0x123abc');
    });

    it('returns correct explorer URL for BSC transactions', () => {
      const txHash = '0x456def';
      expect(getExplorerTxUrl(txHash, 'bsc')).toBe('https://bscscan.com/tx/0x456def');
    });

    it('returns correct explorer URL for Polygon transactions', () => {
      const txHash = '0x789ghi';
      expect(getExplorerTxUrl(txHash, 'polygon')).toBe('https://polygonscan.com/tx/0x789ghi');
    });

    it('returns correct explorer URL for Solana transactions', () => {
      const txHash = 'abc123xyz';
      expect(getExplorerTxUrl(txHash, 'solana')).toBe('https://solscan.io/tx/abc123xyz');
    });

    it('defaults to etherscan for unknown chains', () => {
      const txHash = '0x123abc';
      expect(getExplorerTxUrl(txHash, 'unknown')).toBe('https://etherscan.io/tx/0x123abc');
    });
  });

  describe('getExplorerAddressUrl', () => {
    it('returns correct explorer URL for ethereum addresses', () => {
      const address = '0x1234567890abcdef1234567890abcdef12345678';
      expect(getExplorerAddressUrl(address, 'ethereum')).toBe(`https://etherscan.io/address/${address}`);
    });

    it('returns correct explorer URL for BSC addresses', () => {
      const address = '0x1234567890abcdef1234567890abcdef12345678';
      expect(getExplorerAddressUrl(address, 'bsc')).toBe(`https://bscscan.com/address/${address}`);
    });

    it('returns correct explorer URL for Solana addresses', () => {
      const address = 'DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK';
      expect(getExplorerAddressUrl(address, 'solana')).toBe(`https://solscan.io/account/${address}`);
    });

    it('handles case-insensitive chain names', () => {
      const address = '0x1234567890abcdef1234567890abcdef12345678';
      expect(getExplorerAddressUrl(address, 'ETHEREUM')).toBe(`https://etherscan.io/address/${address}`);
    });
  });
});
