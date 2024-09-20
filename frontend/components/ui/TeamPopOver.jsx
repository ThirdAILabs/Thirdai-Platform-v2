import { useState } from 'react';
import { alpha } from '@mui/material/styles';
import { Typography, MenuItem, Button } from '@mui/material';
import Scrollbar from '../../components/scrollbar';
import MenuPopover from '../../components/menu-popover';

const myTeam = [
  { id: 1, name: 'Tharun Medini' },
  { id: 2, name: 'Benito Geordie' },
  { id: 3, name: 'Sid Jain' }
];
const ITEM_HEIGHT = 64;

export default function TeamPopover({ selectedTeam, setSelectedTeam }) {
  const [openPopover, setOpenPopover] = useState(null);

  const handleOpenPopover = (event) => {
    setOpenPopover(event.currentTarget);
  };

  const handleClosePopover = () => {
    setOpenPopover(null);
  };

  const handleTeamSelect = (teamName) => {
    setSelectedTeam(teamName); // Set the selected team
    handleClosePopover(); // Close the popover after selecting
  };

  return (
    <>
      <Button
        sx={{
          bgcolor: (theme) =>
            alpha(
              theme.palette.secondary.main,
              theme.palette.action.selectedOpacity
            ),
          color: (theme) => theme.palette.secondary.dark,
          px: 2
        }}
        onClick={handleOpenPopover}
      >
        {selectedTeam ? selectedTeam : 'Select Team'}
      </Button>

      <MenuPopover
        open={openPopover}
        onClose={handleClosePopover}
        sx={{ width: 320 }}
      >
        <Typography variant="h6" sx={{ p: 1.5 }}>
          My Team <Typography component="span">({myTeam.length})</Typography>
        </Typography>

        <Scrollbar sx={{ height: ITEM_HEIGHT * 5 }}>
          {myTeam.map((item) => (
            <MenuItem
              key={item?.id}
              sx={{ height: ITEM_HEIGHT }}
              onClick={() => handleTeamSelect(item.name)} // Select the team
            >
              <Typography variant="subtitle2">{item.name}</Typography>
            </MenuItem>
          ))}
        </Scrollbar>
      </MenuPopover>
    </>
  );
}
