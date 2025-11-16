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
  if (loading) {
    return (
      <Typography variant="body2" color="text.secondary">
        Loading file details…
      </Typography>
    );
  }

  if (error) {
    return (
      <Typography variant="body2" color="error">
        {error}
      </Typography>
    );
  }

  return (
    <Card>
      <CardContent>
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
