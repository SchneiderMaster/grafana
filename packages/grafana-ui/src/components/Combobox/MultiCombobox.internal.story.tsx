import { action } from '@storybook/addon-actions';
import { useArgs } from '@storybook/preview-api';
import type { Meta, StoryObj } from '@storybook/react';

import { MultiCombobox } from './MultiCombobox';

const meta: Meta<typeof MultiCombobox> = {
  title: 'Forms/MultiCombobox',
  component: MultiCombobox,
};

const commonArgs = {
  options: [
    { label: 'Option 1', value: 'option1' },
    { label: 'Option 2', value: 'option2' },
    { label: 'Option 3', value: 'option3' },
    { label: 'Option 4', value: 'option4' },
    { label: 'Option 5', value: 'option5' },
    { label: 'S', value: 'option6' },
    { label: 'Medium', value: 'option7' },
    { label: 'Laaaaaaaaaaaarge', value: 'option8' },
    { label: 'Big options with a lot of text', value: 'option9' },
    { label: 'Option 10', value: 'option10' },
  ],
  value: ['option2'],
  placeholder: 'Select multiple options...',
};

export default meta;

type Story = StoryObj<typeof MultiCombobox>;

export const Basic: Story = {
  args: commonArgs,
  render: (args) => {
    const [{ value }, setArgs] = useArgs();

    return (
      <MultiCombobox
        {...args}
        value={value}
        onChange={(val) => {
          action('onChange')(val);
          setArgs({ value: val });
        }}
      />
    );
  },
};
