import { CliServices } from "../../../utilities/shared/services";
import { GlobalOptions } from "../../../utilities/shared/global-options";

export interface ApiMetadataOptions {
  data?: string;
  file?: string;
  object?: string;
  view?: string;
  pageLayout?: string;
  pageLayoutTab?: string;
  pageLayoutType?: string;
}

export interface ApiMetadataContext {
  type: string;
  operation: string;
  arg?: string;
  options: ApiMetadataOptions;
  services: CliServices;
  globalOptions: GlobalOptions;
}
