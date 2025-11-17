import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';

export interface FileMetadataField {
  label: string;
  value?: string | number;
}

export interface FileDetailsPanelProps {
  title: string;
  subtitle?: string;
  metadata: FileMetadataField[];
  usage?: FileMetadataField[];
  loading?: boolean;
  error?: string;
}

export function FileDetailsPanel({ title, subtitle, metadata, usage = [], loading, error }: FileDetailsPanelProps) {
  const wrapperStyle = {
    backgroundColor: '#e3f2fd',
    display: 'inline-flex',
    flexDirection: 'column',
    minHeight: 'max-content',
    minWidth: 'max-content',
  } as const;

  if (loading) {
    return (
        <Box sx={wrapperStyle}>
          <Typography variant="body2" color="text.secondary">
            Loading file details…
          </Typography>
        </Box>
    );
  }

  if (error) {
    return (
        <Box sx={wrapperStyle}>
          <Typography variant="body2" color="error">
            {error}
          </Typography>
        </Box>
    );
  }

  return (
    <Box sx={wrapperStyle}>
      <Card sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <CardContent sx={{ flex: 1, overflow: 'auto' }}>
          <Typography variant="h5" gutterBottom>
            {title}
          </Typography>
          {subtitle ? (
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              {subtitle}
            </Typography>
          ) : null}
          <Section heading="Metadata" fields={metadata} />
          {usage.length ? <Section heading="Usage" fields={usage} /> : null}
        </CardContent>
      </Card>
    </Box>
  );
}

interface SectionProps {
  heading: string;
  fields: FileMetadataField[];
}

function Section({ heading, fields }: SectionProps) {
  if (!fields.length) return null;
  return (
    <Box mt={3}>
      <Typography variant="subtitle1" gutterBottom>
        {heading}
      </Typography>
      <List dense>
        {fields.map((field) => (
          <ListItem key={field.label} disableGutters>
            <ListItemText primary={field.label} secondary={field.value ?? '—'} />
          </ListItem>
        ))}
      </List>
    </Box>
  );
}
