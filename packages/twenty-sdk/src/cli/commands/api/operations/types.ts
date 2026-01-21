import { CliServices } from '../../../utilities/shared/services';
import { GlobalOptions } from '../../../utilities/shared/global-options';

export interface ApiCommandOptions {
  limit?: string;
  all?: boolean;
  filter?: string;
  include?: string;
  cursor?: string;
  sort?: string;
  order?: string;
  fields?: string;
  param?: string[];
  data?: string;
  file?: string;
  set?: string[];
  force?: boolean;
  yes?: boolean;
  ids?: string;
  format?: string;
  output?: string;
  outputFile?: string;
  batchSize?: string;
  dryRun?: boolean;
  continueOnError?: boolean;
  field?: string;
  fieldsList?: string;
  source?: string;
  target?: string;
  priority?: string;
}

export interface ApiOperationContext {
  object: string;
  arg?: string;
  arg2?: string;
  options: ApiCommandOptions;
  services: CliServices;
  globalOptions: GlobalOptions;
}
