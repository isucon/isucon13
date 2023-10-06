import Button from '@mui/joy/Button';
import Divider from '@mui/joy/Divider';
import FormControl from '@mui/joy/FormControl';
import FormLabel from '@mui/joy/FormLabel';
import Input from '@mui/joy/Input';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { useForm } from 'react-hook-form';
import { apiClient } from '~/api/client';

interface FormValues {
  name: string;
  display_name: string;
  description: string;
  password: string;
}

export default function SignupPage(): React.ReactElement {
  const form = useForm<FormValues>();

  const onSubmit = async (data: FormValues) => {
    console.log(data);
    const res = await apiClient.post$user({
      requestBody: {
        name: data.name,
        display_name: data.display_name,
        description: data.description,
        password: data.password,
      },
    });
    console.log(res);
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <Stack maxWidth="500px" mx="auto" my={5} gap={3}>
        <FormControl>
          <FormLabel>Username</FormLabel>
          <Input {...form.register('display_name')} required />
        </FormControl>
        <FormControl>
          <FormLabel>Description</FormLabel>
          <Input {...form.register('description')} required />
        </FormControl>
        <Divider />
        <FormControl>
          <FormLabel>Login ID</FormLabel>
          <Input {...form.register('name')} required />
        </FormControl>
        <FormControl>
          <FormLabel>Password</FormLabel>
          <Input {...form.register('password')} required />
        </FormControl>
        <FormControl>
          <Button type="submit">登録</Button>
        </FormControl>
      </Stack>
    </form>
  );
}
