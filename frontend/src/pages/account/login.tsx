import Button from '@mui/joy/Button';
import Divider from '@mui/joy/Divider';
import FormControl from '@mui/joy/FormControl';
import FormLabel from '@mui/joy/FormLabel';
import Input from '@mui/joy/Input';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { apiClient } from '~/api/client';
import { useUserMe } from '~/api/hooks';
import { useGlobalToastQueue } from '~/components/toast/toast';

interface FormValues {
  username: string;
  password: string;
}

export default function AccountPage(): React.ReactElement {
  const form = useForm<FormValues>();
  const toast = useGlobalToastQueue();
  const navigate = useNavigate();

  const userMe = useUserMe({
    revalidateIfStale: false,
    revalidateOnFocus: false,
    revalidateOnMount: false,
    revalidateOnReconnect: false,
  });

  const onSubmit = async (data: FormValues) => {
    try {
      const res = await apiClient.post$login({
        requestBody: {
          username: data.username,
          password: data.password,
        },
      });
      console.log(res);
      toast.add(
        {
          type: 'success',
          title: 'Login Success',
          message: 'ログインに成功しました',
        },
        { timeout: 3000 },
      );
      await Promise.all([userMe.mutate()]);
      navigate('/');
    } catch (e) {
      console.warn(e);
      toast.add(
        {
          type: 'error',
          title: 'Login Failed',
          message: 'ログインに失敗しました',
        },
        { timeout: 3000 },
      );
    }
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <Stack maxWidth="500px" mx="auto" my={5} gap={3}>
        <FormControl>
          <FormLabel>Username</FormLabel>
          <Input {...form.register('username')} required />
        </FormControl>
        <FormControl>
          <FormLabel>Password</FormLabel>
          <Input {...form.register('password')} required />
        </FormControl>
        <FormControl>
          <Button type="submit">ログイン</Button>
        </FormControl>
        <Divider />
        <FormControl>
          <Button component={Link} to="/account/signup" variant="plain">
            新規登録
          </Button>
        </FormControl>
      </Stack>
    </form>
  );
}
