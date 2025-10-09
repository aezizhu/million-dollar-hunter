import {
  formatCurrency,
  formatNumber,
  formatPercent,
  formatAddress,
  isValidEthAddress,
  isValidSolanaAddress,
  isValidAddress,
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
});
