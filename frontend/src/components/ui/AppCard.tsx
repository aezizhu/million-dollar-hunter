import { Card, CardProps } from '@mui/material';

export function AppCard(props: CardProps) {
  return (
    <Card
      elevation={0}
      sx={{
        border: '1px solid',
        borderColor: 'divider',
        ...props.sx,
      }}
      {...props}
    />
  );
}
