import Box from '@mui/joy/Box';
import Button from '@mui/joy/Button';
import DialogContent from '@mui/joy/DialogContent';
import DialogTitle from '@mui/joy/DialogTitle';
import Modal from '@mui/joy/Modal';
import ModalDialog from '@mui/joy/ModalDialog';
import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { normalizeUrl } from '~/api/url';

export interface LoginModalProps {
  isOpen: boolean;
  onClose: () => void;
}
export function LoginModal(props: LoginModalProps): React.ReactElement {
  const location = useLocation();

  switch (location.pathname) {
    case '/account/login':
    case '/account/signup':
      // not show login modal in login page and signup page
      return <></>;
  }

  return (
    <>
      <Modal open={props.isOpen} onClose={props.onClose}>
        <ModalDialog size="lg" sx={{ minWidth: '600px' }}>
          <DialogTitle>ログインが必要です</DialogTitle>
          <DialogContent>
            <Box>
              <Button
                component={Link}
                to={normalizeUrl(`/account/login`)}
                variant="soft"
                onClick={() => props.onClose()}
              >
                ログインページへ移動
              </Button>
            </Box>
          </DialogContent>
        </ModalDialog>
      </Modal>
    </>
  );
}
