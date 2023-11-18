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
import { useGlobalToastQueue } from '../toast/toast';

export interface ChangeIconDialogProps {
  isOpen: boolean;
  onClose?(): void;
  onSubmit?(iconBase64: string): Promise<void> | void;
}
export function ChangeIconDialog(
  props: ChangeIconDialogProps,
): React.ReactElement {
  const inputRef = React.useRef<HTMLDivElement>(null);
  const [isLoading, setIsLoading] = React.useState<boolean>(false);
  const toast = useGlobalToastQueue();
  const submitHandler = React.useCallback(async () => {
    if (props.onSubmit) {
      setIsLoading(true);
      try {
        const file = inputRef.current?.querySelector('input')?.files?.[0];
        if (!file) {
          return;
        }
        if (file.type !== 'image/jpeg') {
          toast.add({
            type: 'error',
            title: '画像形式エラー',
            message: '画像ファイルはjpeg形式のみ対応しています',
          });
          return;
        }
        const promise = new Promise<string | undefined>((resolve) => {
          const reader = new FileReader();
          reader.onload = (e) => {
            if (!e.currentTarget) {
              resolve(undefined);
              return;
            }
            const base64 = (e.currentTarget as any).result as string;
            resolve(base64);
          };
          reader.readAsDataURL(file);
        });
        const text = await promise;

        const prefix = `data:image/jpeg;base64,`;
        await props.onSubmit(text?.slice(prefix.length) ?? '');
        props.onClose?.();
      } finally {
        setIsLoading(false);
      }
    }
  }, [props.onSubmit, props.onClose]);

  return (
    <>
      <Modal open={props.isOpen} onClose={props.onClose}>
        <ModalDialog size="lg" sx={{ minWidth: '600px' }}>
          <DialogTitle>アイコン変更</DialogTitle>
          <DialogContent>
            アイコンの画像ファイルを指定してください。
          </DialogContent>
          <form onSubmit={submitHandler}>
            <Stack spacing={2}>
              <FormControl>
                <FormLabel>アイコン画像ファイル Jpeg画像のみ対応</FormLabel>
                <Input ref={inputRef} type="file" required sx={{ p: 1 }} />
              </FormControl>
              <Button onClick={submitHandler} loading={isLoading}>
                変更
              </Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </>
  );
}
