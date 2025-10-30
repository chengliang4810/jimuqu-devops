declare namespace Api {
  /**
   * namespace Host
   *
   * backend api module: "hosts"
   */
  namespace Host {
    type CommonSearchParams = Pick<Common.PaginatingCommonParams, 'current' | 'size'>;

    /** Host connection status */
    type HostStatus = '未连接' | '正常' | '异常';

    /** Host */
    type Host = Common.CommonRecord<{
      /** Host name */
      name: string;
      /** Host description */
      description: string | null;
      /** Host address (IP or domain) */
      host: string;
      /** SSH port */
      port: number;
      /** SSH username */
      username: string;
      /** SSH private key path */
      ssh_key_path: string | null;
      /** SSH password (encrypted) */
      ssh_password: string | null;
      /** Host tags (comma separated) */
      tags: string | null;
      /** Host group */
      group: string | null;
      /** Connection status */
      status: HostStatus;
      /** Last connected time */
      last_connected_at: string | null;
      /** Operating system type */
      os_type: string | null;
      /** Operating system version */
      os_version: string | null;
      /** Whether the host is active */
      is_active: boolean;
    }>;

    /** Host search params */
    type HostSearchParams = CommonType.RecordNullable<
      {
        search: string;
        group: string;
        status: HostStatus;
        is_active: boolean;
      } & CommonSearchParams
    >;

    /** Host list response */
    type HostListResponse = {
      hosts: Host[];
      total: number;
      page: number;
      size: number;
      pages: number;
    };

    /** Host create data */
    type HostCreate = Pick<
      Host,
      'name' | 'description' | 'host' | 'port' | 'username' | 'ssh_key_path' | 'ssh_password' | 'tags' | 'group'
    >;

    /** Host update data */
    type HostUpdate = Partial<HostCreate> & {
      is_active?: boolean;
    };

    /** Host connection test result */
    type HostTestConnection = {
      success: boolean;
      message: string;
      response_time: number | null;
      os_info: {
        uname: string;
        user: string | null;
        os_type: string | null;
        os_version: string | null;
      } | null;
    };
  }
}