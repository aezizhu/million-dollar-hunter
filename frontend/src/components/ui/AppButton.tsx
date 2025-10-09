import { Button, ButtonProps } from '@mui/material';

export function AppButton(props: ButtonProps) {
  return (
    <Button
      variant="contained"
      size="medium"
      {...props}
    />
  );
}
