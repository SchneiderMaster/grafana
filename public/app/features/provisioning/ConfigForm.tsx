import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom-v5-compat';

import { AppEvents } from '@grafana/data';
import { getAppEvents } from '@grafana/runtime';
import {
  Button,
  Combobox,
  ComboboxOption,
  Field,
  FieldSet,
  Input,
  MultiCombobox,
  RadioButtonGroup,
  SecretInput,
  Stack,
  Switch,
} from '@grafana/ui';
import { FormPrompt } from 'app/core/components/FormPrompt/FormPrompt';

import { TokenPermissionsInfo } from './TokenPermissionsInfo';
import { Repository, RepositorySpec } from './api';
import { useCreateOrUpdateRepository } from './hooks';
import { RepositoryFormData, WorkflowOption } from './types';
import { dataToSpec, specToData } from './utils/data';

const typeOptions = ['GitHub', 'Local', 'S3'].map((label) => ({ label, value: label.toLowerCase() }));
const targetOptions = [
  { value: 'instance', label: 'Entire instance' },
  { value: 'folder', label: 'Managed folder' },
];

const workflowOptions: Array<ComboboxOption<WorkflowOption>> = [
  { label: 'Push', value: 'push' },
  { label: 'Branch', value: 'branch' },
];

const appEvents = getAppEvents();

function getDefaultValues(repository?: RepositorySpec): RepositoryFormData {
  if (!repository) {
    return {
      type: 'github',
      title: '',
      token: '',
      repositoryUrl: '',
      branch: 'main',
      generateDashboardPreviews: true,
      workflows: ['push', 'branch'],
      sync: {
        enabled: false,
        target: 'instance',
        intervalSeconds: 60,
      },
      readOnly: false,
    };
  }
  return specToData(repository);
}

export interface ConfigFormProps {
  data?: Repository;
}
export function ConfigForm({ data }: ConfigFormProps) {
  const [submitData, request] = useCreateOrUpdateRepository(data?.metadata?.name);
  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isDirty },
    setValue,
    watch,
    getValues,
  } = useForm<RepositoryFormData>({ defaultValues: getDefaultValues(data?.spec) });
  const isEdit = Boolean(data?.metadata?.name);
  const [tokenConfigured, setTokenConfigured] = useState(isEdit);
  const navigate = useNavigate();
  const type = watch('type');

  useEffect(() => {
    if (request.isSuccess) {
      const formData = getValues();

      appEvents.publish({
        type: AppEvents.alertSuccess.name,
        payload: ['Repository settings saved'],
      });
      reset(formData);
      setTimeout(() => {
        navigate('/admin/provisioning');
      }, 300);
    }
  }, [request.isSuccess, reset, getValues, navigate]);

  const onSubmit = (data: RepositoryFormData) => {
    const spec = dataToSpec(data);
    submitData(spec);
  };

  // NOTE: We do not want the lint option to be listed.
  return (
    <form onSubmit={handleSubmit(onSubmit)} style={{ maxWidth: 700 }}>
      <FormPrompt onDiscard={reset} confirmRedirect={isDirty} />
      <Field label={'Repository type'}>
        <Controller
          name={'type'}
          control={control}
          render={({ field: { ref, onChange, ...field } }) => {
            return (
              <Combobox
                options={typeOptions}
                onChange={(value) => onChange(value?.value)}
                placeholder={'Select repository type'}
                disabled={!!data?.spec}
                {...field}
              />
            );
          }}
        />
      </Field>
      <Field
        label={'Title'}
        description={'A human-readable name for the config'}
        invalid={!!errors.title}
        error={errors?.title?.message}
      >
        <Input {...register('title', { required: 'This field is required.' })} placeholder={'My config'} />
      </Field>
      {type === 'github' && (
        <FieldSet label="GitHub">
          <Field label={'Token'} required error={errors?.token?.message} invalid={!!errors.token}>
            <Controller
              name={'token'}
              control={control}
              rules={{ required: isEdit ? false : 'This field is required.' }}
              render={({ field: { ref, ...field } }) => {
                return (
                  <SecretInput
                    {...field}
                    id={'token'}
                    placeholder={'ghp_yourTokenHere1234567890abcdEFGHijklMNOP'}
                    isConfigured={tokenConfigured}
                    onReset={() => {
                      setValue('token', '');
                      setTokenConfigured(false);
                    }}
                  />
                );
              }}
            />
          </Field>
          <TokenPermissionsInfo />
          <Field
            label={'Repository URL'}
            error={errors?.repositoryUrl?.message}
            invalid={!!errors?.repositoryUrl}
            description={'Enter the GitHub repository URL or owner/repository name'}
            required
          >
            <Input
              {...register('repositoryUrl', {
                required: 'This field is required.',
                pattern: {
                  value: /^(?:https:\/\/github\.com\/)?[^/]+\/[^/]+$/,
                  message: 'Please enter a valid GitHub repository URL or owner/repository (e.g. grafana/grafana)',
                },
              })}
              placeholder={'https://github.com/username/repo-name or username/repo-name'}
            />
          </Field>
          <Field label={'Branch'}>
            <Input {...register('branch')} placeholder={'main'} />
          </Field>
          <Field label={'Workflows'} required error={errors?.workflows?.message} invalid={!!errors?.workflows}>
            <Controller
              name={'workflows'}
              control={control}
              rules={{ required: 'This field is required.' }}
              render={({ field: { ref, ...field } }) => (
                <MultiCombobox options={workflowOptions} placeholder={'Select workflows'} {...field} />
              )}
            />
          </Field>
          <Field label={'Show dashboard previews'}>
            <Switch {...register('generateDashboardPreviews')} id={'generateDashboardPreviews'} />
          </Field>
          {/* The lint option is intentionally not presented here, as it's an experimental feature. */}
        </FieldSet>
      )}

      {type === 'local' && (
        <Field label={'Local path'} error={errors?.path?.message} invalid={!!errors?.path}>
          <Input {...register('path', { required: 'This field is required.' })} placeholder={'/path/to/repo'} />
        </Field>
      )}

      {type === 's3' && (
        <FieldSet label="local">
          <Field label={'S3 bucket'} error={errors?.bucket?.message} invalid={!!errors?.bucket}>
            <Input {...register('bucket', { required: 'This field is required.' })} placeholder={'bucket-name'} />
          </Field>
          <Field label={'S3 region'} error={errors?.region?.message} invalid={!!errors?.region}>
            <Input {...register('region', { required: 'This field is required.' })} placeholder={'us-west-2'} />
          </Field>
        </FieldSet>
      )}
      <FieldSet label="Sync Settings">
        <Field label={'Enabled'} description={'Once sync is enabled, the target cannot be changed.'}>
          <Switch {...register('sync.enabled')} id={'sync.enabled'} />
        </Field>
        <Field label={'Target'} required error={errors?.sync?.target?.message} invalid={!!errors?.sync?.target}>
          <Controller
            name={'sync.target'}
            control={control}
            rules={{ required: 'This field is required.' }}
            render={({ field: { ref, onChange, ...field } }) => {
              return (
                <RadioButtonGroup
                  options={targetOptions}
                  onChange={onChange}
                  disabled={Boolean(data?.status?.sync.state)}
                  {...field}
                />
              );
            }}
          />
        </Field>
        <Field label={'Interval (seconds)'}>
          <Input {...register('sync.intervalSeconds')} type={'number'} placeholder={'60'} />
        </Field>
      </FieldSet>
      <FieldSet label="Advanced Settings">
        <Field label={'Read Only'} description={'Disable writing to this repository'}>
          <Switch {...register('readOnly')} id={'readOnly'} />
        </Field>
      </FieldSet>
      <Stack gap={2}>
        <Button type={'submit'} disabled={request.isLoading}>
          {request.isLoading ? 'Saving...' : 'Save'}
        </Button>
      </Stack>
    </form>
  );
}
