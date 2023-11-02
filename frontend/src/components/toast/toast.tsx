import Alert, { type AlertProps } from '@mui/joy/Alert';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import {
  ToastQueue,
  type ToastState,
  useToastQueue,
  type QueuedToast,
} from '@react-stately/toast';
import React from 'react';

export const toastQueue = new ToastQueue<ToastItem>({
  maxVisibleToasts: 5,
});

export interface ToastItem {
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
}

export function useGlobalToastQueue(): ToastState<ToastItem> {
  return useToastQueue(toastQueue);
}

export function Toast(): React.ReactElement {
  const state = useGlobalToastQueue();

  return (
    <Stack
      sx={{ position: 'fixed', bottom: 0, right: 0, minWidth: '300px' }}
      gap={2}
      m={2}
    >
      {state.visibleToasts.map((toast) => (
        <ToastItem key={toast.key} toast={toast} />
      ))}
    </Stack>
  );
}

export function ToastItem({
  toast,
}: {
  toast: QueuedToast<ToastItem>;
}): React.ReactElement {
  React.useEffect(() => {
    if (!toast.timer || !toast.timeout) {
      return;
    }
    toast.timer.reset(toast.timeout);
    return () => {
      toast.timer?.pause();
    };
  });
  return (
    <Alert key={toast.key} color={colorMap(toast.content.type)}>
      <div>
        <div>{toast.content.title}</div>
        <Typography level="body-sm" color={colorMap(toast.content.type)}>
          {toast.content.message}
        </Typography>
      </div>
    </Alert>
  );
}

function colorMap(type: ToastItem['type']): AlertProps['color'] {
  switch (type) {
    case 'info':
      return 'neutral';
    case 'success':
      return 'success';
    case 'warning':
      return 'warning';
    case 'error':
      return 'danger';
  }
}
