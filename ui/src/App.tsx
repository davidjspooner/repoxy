import { Box, Typography } from '@mui/material';
import './app.css';

function App() {
  return (
    <Box
      className="app-shell"
      component="main"
      display="flex"
      alignItems="center"
      justifyContent="center"
      height="100%"
    >
      <Typography
        className="todo-text"
        variant="h2"
        component="p"
        letterSpacing={4}
        textTransform="uppercase"
      >
        TODO
      </Typography>
    </Box>
  );
}

export default App;
