import styled from '@emotion/styled';
import Dropdown from '@mui/joy/Dropdown';
import Menu from '@mui/joy/Menu';
import MenuButton from '@mui/joy/MenuButton';
import MenuItem from '@mui/joy/MenuItem';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { MdChair } from 'react-icons/md';
import { Link } from 'react-router-dom';

interface LayoutProps {
  children: React.ReactNode;
}
export function Layout(props: LayoutProps): React.ReactElement {
  return (
    <div>
      <Stack
        direction="row"
        justifyContent="space-between"
        alignItems="center"
        spacing={2}
        sx={{
          px: 3,
          py: 1,
          borderBottom: (t) => `1px solid ${t.vars.palette.background.level1}`,
        }}
      >
        <LogoLink to="/">
          <MdChair size="25px" />
          <span>ISUTUBE</span>
        </LogoLink>
        <div>
          <Dropdown>
            <MenuButton>account</MenuButton>
            <Menu>
              <MenuItem>aaa</MenuItem>
              <MenuItem>bbb</MenuItem>
            </Menu>
          </Dropdown>
        </div>
      </Stack>
      {props.children}
    </div>
  );
}

const LogoLink = styled(Link)`
  display: flex;
  align-items: flex-end;
  text-decoration: none;
  color: inherit;
  font-size: 1.3rem;
  font-weight: 700;
  svg {
    margin-right: 0.2rem;
    color: #ff0000;
  }
`;
