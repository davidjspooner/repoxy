import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

/**
 * Conceptual root panel; in practice the application immediately navigates
 * to the RepositoryTypesPanel. Exposed here for completeness/testing.
 */
export function RootPanel() {
  return (
    <Box display="flex" alignItems="center" justifyContent="center" minHeight="200px">
      <Typography variant="body1" color="text.secondary">
        Loading repository types…
      </Typography>
    </Box>
  );
}
