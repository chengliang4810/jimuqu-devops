<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { jsonClone } from '@sa/utils';
import { fetchCreateHost, fetchUpdateHost } from '@/service/api';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'HostOperateDrawer'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.Host.Host | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.host.addHost'),
    edit: $t('page.host.editHost')
  };
  return titles[props.operateType];
});

type Model = Pick<
  Api.Host.Host,
  'name' | 'description' | 'host' | 'port' | 'username' | 'ssh_key_path' | 'ssh_password' | 'tags' | 'group'
>;

const model = ref(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    description: '',
    host: '',
    port: 22,
    username: '',
    ssh_key_path: '',
    ssh_password: '',
    tags: '',
    group: ''
  };
}

type RuleKey = Extract<keyof Model, 'name' | 'host' | 'port' | 'username'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  host: defaultRequiredRule,
  port: {
    ...defaultRequiredRule,
    type: 'number',
    message: $t('page.host.form.portInvalid')
  },
  username: defaultRequiredRule
};

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, jsonClone(props.rowData));
  }
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  try {
    if (props.operateType === 'add') {
      await fetchCreateHost(model.value);
      window.$message?.success($t('common.addSuccess'));
    } else {
      await fetchUpdateHost(props.rowData!.id, model.value);
      window.$message?.success($t('common.updateSuccess'));
    }

    closeDrawer();
    emit('submitted');
  } catch (error) {
    window.$message?.error($t('common.requestFailed'));
  }
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
  }
});
</script>

<template>
  <NDrawer v-model:show="visible" display-directive="show" :width="480">
    <NDrawerContent :title="title" :native-scrollbar="false" closable>
      <NForm ref="formRef" :model="model" :rules="rules">
        <NFormItem :label="$t('page.host.name')" path="name">
          <NInput
            v-model:value="model.name"
            :placeholder="$t('page.host.form.name')"
            maxlength="100"
            show-count
          />
        </NFormItem>

        <NFormItem :label="$t('page.host.description')" path="description">
          <NInput
            v-model:value="model.description"
            type="textarea"
            :placeholder="$t('page.host.form.description')"
            :autosize="{ minRows: 2, maxRows: 4 }"
            maxlength="500"
            show-count
          />
        </NFormItem>

        <div class="grid grid-cols-2 gap-16px">
          <NFormItem :label="$t('page.host.host')" path="host">
            <NInput
              v-model:value="model.host"
              :placeholder="$t('page.host.form.host')"
              maxlength="100"
            />
          </NFormItem>

          <NFormItem :label="$t('page.host.port')" path="port">
            <NInputNumber
              v-model:value="model.port"
              :placeholder="$t('page.host.form.port')"
              :min="1"
              :max="65535"
              class="w-full"
            />
          </NFormItem>
        </div>

        <NFormItem :label="$t('page.host.username')" path="username">
          <NInput
            v-model:value="model.username"
            :placeholder="$t('page.host.form.username')"
            maxlength="50"
          />
        </NFormItem>

        <NDivider>{{ $t('page.host.auth.title') }}</NDivider>

        <NFormItem :label="$t('page.host.auth.keyPath')" path="ssh_key_path">
          <NInput
            v-model:value="model.ssh_key_path"
            :placeholder="$t('page.host.auth.keyPathPlaceholder')"
            maxlength="500"
          />
        </NFormItem>

        <NFormItem :label="$t('page.host.auth.password')" path="ssh_password">
          <NInput
            v-model:value="model.ssh_password"
            type="password"
            show-password-on="click"
            :placeholder="$t('page.host.auth.passwordPlaceholder')"
            maxlength="200"
          />
        </NFormItem>

        <NDivider>{{ $t('page.host.organization.title') }}</NDivider>

        <div class="grid grid-cols-2 gap-16px">
          <NFormItem :label="$t('page.host.group')" path="group">
            <NInput
              v-model:value="model.group"
              :placeholder="$t('page.host.form.group')"
              maxlength="50"
            />
          </NFormItem>

          <NFormItem :label="$t('page.host.tags')" path="tags">
            <NInput
              v-model:value="model.tags"
              :placeholder="$t('page.host.form.tags')"
              maxlength="200"
            />
          </NFormItem>
        </div>

        <NAlert type="info" :show-icon="false" class="mt-16px">
          <template #icon>
            <NIcon size="16">
              <svg viewBox="0 0 24 24">
                <path fill="currentColor" d="M13 9h-2V7h2m0 10h-2v-6h2m-1-9A10 10 0 0 0 2 12a10 10 0 0 0 10 10a10 10 0 0 0 10-10A10 10 0 0 0 12 2Z"/>
              </svg>
            </NIcon>
          </template>
          <span>{{ $t('page.host.form.tips') }}</span>
        </NAlert>
      </NForm>

      <template #footer>
        <NSpace :size="16">
          <NButton @click="closeDrawer">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped></style>