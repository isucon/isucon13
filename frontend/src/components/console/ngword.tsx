import Button from '@mui/joy/Button';
import DialogContent from '@mui/joy/DialogContent';
import DialogTitle from '@mui/joy/DialogTitle';
import FormControl from '@mui/joy/FormControl';
import FormLabel from '@mui/joy/FormLabel';
import Input from '@mui/joy/Input';
import Modal from '@mui/joy/Modal';
import ModalDialog from '@mui/joy/ModalDialog';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { useForm } from 'react-hook-form';

export interface NewNgWordDialogProps {
  isOpen: boolean;
  onClose?(): void;
  onSubmit?(word: string): Promise<void>;
}
export interface NgWordForm {
  word: string;
}
export function NewNgWordDialog(
  props: NewNgWordDialogProps,
): React.ReactElement {
  const form = useForm<NgWordForm>({
    defaultValues: {
      word: '',
    },
  });
  const [isLoading, setIsLoading] = React.useState<boolean>(false);
  const submitHandler = form.handleSubmit(async (data) => {
    if (props.onSubmit) {
      setIsLoading(true);
      try {
        await props.onSubmit(data.word);
      } finally {
        setIsLoading(false);
      }
    }
    form.reset();
  });

  return (
    <>
      <Modal open={props.isOpen} onClose={props.onClose}>
        <ModalDialog size="lg" sx={{ minWidth: '600px' }}>
          <DialogTitle>NGワード登録</DialogTitle>
          <DialogContent>
            NGワードとして指定する内容を入力してください。
          </DialogContent>
          <form onSubmit={submitHandler}>
            <Stack spacing={2}>
              <FormControl>
                <FormLabel>ワード</FormLabel>
                <Input required {...form.register('word')} />
              </FormControl>
              <Button type="submit" loading={isLoading}>
                作成
              </Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </>
  );
}
