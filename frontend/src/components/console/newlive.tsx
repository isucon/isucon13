import Autocomplete from '@mui/joy/Autocomplete';
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
import { Controller, useForm } from 'react-hook-form';
import { useTags } from '~/api/hooks';
import { Schemas } from '~/api/types';

export interface NewLiveFormValue {
  title: string;
  description: string;
  tags: number[];
  startAt: string;
  endAt: string;
}

export interface NewLiveDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit?(form: NewLiveFormValue): Promise<void>;
}
export function NewLiveDialog(props: NewLiveDialogProps): React.ReactElement {
  const [isLoading, setIsLoading] = React.useState<boolean>(false);
  const form = useForm<NewLiveFormValue>({
    defaultValues: {
      title: '',
      description: '',
      startAt: '',
      endAt: '',
      tags: [],
    },
  });

  const tags = useTags();

  const submitHandler = form.handleSubmit(async (data) => {
    if (props.onSubmit) {
      setIsLoading(true);
      try {
        await props.onSubmit(data);
        props.onClose();
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
          <DialogTitle>予約配信の作成</DialogTitle>
          <DialogContent>配信内容を入力してください。</DialogContent>
          <form onSubmit={submitHandler}>
            <Stack spacing={2}>
              <FormControl>
                <FormLabel>タイトル</FormLabel>
                <Input required {...form.register('title')} />
              </FormControl>
              <FormControl>
                <FormLabel>説明文</FormLabel>
                <Textarea
                  required
                  minRows={3}
                  {...form.register('description')}
                />
              </FormControl>
              <FormControl>
                <FormLabel>タグ</FormLabel>
                <Controller
                  control={form.control}
                  name="tags"
                  render={({ field }) => (
                    <Autocomplete
                      options={tags.data?.tags ?? []}
                      getOptionLabel={(option: Schemas.Tag) => option.name}
                      onChange={(_, value) =>
                        field.onChange(value ? [value.id] : [])
                      }
                      value={
                        tags.data?.tags?.find(
                          (tag) => tag.id === field.value?.[0],
                        ) ?? null
                      }
                      loading={tags.isLoading}
                    />
                  )}
                />
              </FormControl>
              <FormControl>
                <FormLabel>開始日時 2024年04月から指定可能</FormLabel>
                <Input
                  type="datetime-local"
                  required
                  {...form.register('startAt')}
                />
              </FormControl>
              <FormControl>
                <FormLabel>終了日時 2025年4月末まで指定可能</FormLabel>
                <Input
                  type="datetime-local"
                  required
                  {...form.register('endAt')}
                />
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
