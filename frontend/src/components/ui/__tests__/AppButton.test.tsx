import { render, screen } from '@testing-library/react';
import { AppButton } from '../AppButton';

describe('AppButton', () => {
  it('renders button with text', () => {
    render(<AppButton>Click me</AppButton>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('applies variant prop', () => {
    render(<AppButton variant="outlined">Test</AppButton>);
    const button = screen.getByText('Test');
    expect(button).toBeInTheDocument();
  });

  it('handles disabled state', () => {
    render(<AppButton disabled>Disabled</AppButton>);
    const button = screen.getByText('Disabled');
    expect(button).toBeDisabled();
  });
});
