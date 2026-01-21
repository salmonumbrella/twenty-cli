import { CliServices } from '../../../utilities/shared/services';
import { GlobalOptions } from '../../../utilities/shared/global-options';

export interface ApiMetadataOptions {
  data?: string;
  file?: string;
  object?: string;
}

export interface ApiMetadataContext {
  type: string;
  operation: string;
  arg?: string;
  options: ApiMetadataOptions;
  services: CliServices;
  globalOptions: GlobalOptions;
}
