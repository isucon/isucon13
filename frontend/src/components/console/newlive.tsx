import Button from '@mui/joy/Button';
import DialogContent from '@mui/joy/DialogContent';
import DialogTitle from '@mui/joy/DialogTitle';
import FormControl from '@mui/joy/FormControl';
import FormLabel from '@mui/joy/FormLabel';
import Input from '@mui/joy/Input';
import Modal from '@mui/joy/Modal';
import ModalDialog from '@mui/joy/ModalDialog';
import Stack from '@mui/joy/Stack';
import Textarea from '@mui/joy/Textarea';
import React from 'react';

export interface NewLiveDialogProps {
  isOpen: boolean;
  onClose: () => void;
}
export function NewLiveDialog(props: NewLiveDialogProps): React.ReactElement {
  return (
    <>
      <Modal open={props.isOpen} onClose={props.onClose}>
        <ModalDialog size="lg" sx={{ minWidth: '600px' }}>
          <DialogTitle>予約配信の作成</DialogTitle>
          <DialogContent>配信内容を入力してください。</DialogContent>
          <form
            onSubmit={(event: React.FormEvent<HTMLFormElement>) => {
              event.preventDefault();
              props.onClose();
            }}
          >
            <Stack spacing={2}>
              <FormControl>
                <FormLabel>タイトル</FormLabel>
                <Input required />
              </FormControl>
              <FormControl>
                <FormLabel>説明文</FormLabel>
                <Textarea required minRows={3} />
              </FormControl>
              <FormControl>
                <FormLabel>開始日時</FormLabel>
                <Input type="datetime-local" required />
              </FormControl>
              <FormControl>
                <FormLabel>終了日時</FormLabel>
                <Input type="datetime-local" required />
              </FormControl>
              <Button type="submit">作成</Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </>
  );
}
