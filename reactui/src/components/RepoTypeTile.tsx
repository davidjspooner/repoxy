import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import Typography from '@mui/material/Typography';
import type { RepoType } from './types';

export interface RepoTypeTileProps {
  repoType: RepoType;
  selected?: boolean;
}

export function RepoTypeTile({ repoType, selected }: RepoTypeTileProps) {
  return (
    <Card
      elevation={selected ? 6 : 1}
      sx={{
        border: selected ? '2px solid #4c8bf5' : '1px solid rgba(0,0,0,0.2)',
        transition: 'border 0.2s ease, box-shadow 0.2s ease',
        display: 'flex',
        flexDirection: 'column',
        minHeight: 150,
      }}
    >
      <CardActionArea onClick={repoType.onSelect} sx={{ height: '100%' }}>
        <CardContent
          sx={{
            display: 'flex',
            flexDirection: 'column',
            height: '100%',
            justifyContent: 'space-between',
            gap: 1,
          }}
        >
          <Typography variant="h6" gutterBottom>
            {repoType.label}
          </Typography>
          {repoType.description ? (
            <Typography variant="body2" color="text.secondary">
              {repoType.description}
            </Typography>
          ) : null}
        </CardContent>
      </CardActionArea>
    </Card>
  );
}
