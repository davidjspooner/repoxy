import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import Typography from '@mui/material/Typography';
import type { RepoType } from './types';

export interface RepoTypeTileProps {
  repoType: RepoType;
}

export function RepoTypeTile({ repoType }: RepoTypeTileProps) {
  return (
    <Card elevation={1}>
      <CardActionArea onClick={repoType.onSelect} sx={{ height: '100%' }}>
        <CardContent>
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
