// This file is auto-generated by @hey-api/openapi-ts

/**
 * AuthRequest schema
 */
export type AuthRequest = {
    password?: string;
    username?: string;
};

/**
 * AuthResetRequest schema
 */
export type AuthResetRequest = {
    current_password?: string;
    new_password: string;
};

/**
 * AuthResponse schema
 */
export type AuthResponse = {
    expire?: number;
    role?: string;
    token?: string;
    username?: string;
};

/**
 * AuthSignupRequest schema
 */
export type AuthSignupRequest = {
    password: string;
    username: string;
};

/**
 * AuthTokenReponse schema
 */
export type AuthTokenReponse = {
    token?: string;
};

/**
 * AuthTokenRequest schema
 */
export type AuthTokenRequest = {
    /**
     * string parsed by time.ParseDuration, examples include: infinite, 8h, 30m, 20s
     */
    expire?: string;
    /**
     * type of model.Role, valid options: disabled, user, admin
     */
    role?: string;
    /**
     * username shown in logs, does not need to be a valid user in the DB
     */
    username?: string;
};

/**
 * BmcImportConfigurationRequest schema
 */
export type BmcImportConfigurationRequest = {
    /**
     * template file relative to templates directory
     */
    file?: string;
    /**
     * options include: NoReboot, Graceful, Forced
     */
    shutdown_type?: string;
};

/**
 * BmcOsPowerBody schema
 */
export type BmcOsPowerBody = {
    /**
     * string of type redfish.BootSourceOverrideTarget. Common options include: None, Pxe, BiosSetup, Utilities, Diags
     */
    boot_option?: string;
    /**
     * string of type redfish.ResetType. Common options include: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle
     */
    power_option?: string;
};

/**
 * BootImage schema
 */
export type BootImage = {
    butane?: string;
    cmdline?: string;
    id?: (number) | null;
    initrd?: Array<(string)>;
    kernel: string;
    liveimg?: string;
    name: string;
    provision_template?: string;
    provision_templates?: {
        [key: string]: ((string) | null);
    } | null;
    uid?: (string) | null;
    user_data?: string;
    verify?: boolean;
};

/**
 * BootImageAddRequest schema
 */
export type BootImageAddRequest = {
    boot_images?: Array<({
        butane?: string;
        cmdline?: string;
        id?: (number) | null;
        initrd?: Array<(string)>;
        kernel?: string;
        liveimg?: string;
        name?: string;
        provision_template?: string;
        provision_templates?: {
            [key: string]: ((string) | null);
        } | null;
        uid?: (string) | null;
        user_data?: string;
        verify?: boolean;
    } | null)>;
};

/**
 * DataDump schema
 */
export type DataDump = {
    Hosts?: Array<({
        bonds?: Array<({
            bmc?: boolean;
            fqdn?: string;
            id?: (number) | null;
            ifname?: string;
            ip?: string;
            mac?: string;
            mtu?: number;
            peers?: Array<(string)>;
            vlan?: string;
        } | null)>;
        boot_image?: string;
        firmware?: string;
        id?: (number) | null;
        interfaces?: Array<({
            bmc?: boolean;
            fqdn?: string;
            id?: (number) | null;
            ifname?: string;
            ip?: string;
            mac?: string;
            mtu?: number;
            vlan?: string;
        } | null)>;
        name?: string;
        provision?: boolean;
        tags?: Array<(string)>;
        uid?: (string) | null;
    } | null)>;
    Images?: Array<({
        butane?: string;
        cmdline?: string;
        id?: (number) | null;
        initrd?: Array<(string)>;
        kernel?: string;
        liveimg?: string;
        name?: string;
        provision_template?: string;
        provision_templates?: {
            [key: string]: ((string) | null);
        } | null;
        uid?: (string) | null;
        user_data?: string;
        verify?: boolean;
    } | null)>;
    Users?: Array<{
        created_at?: string;
        enabled?: boolean;
        hash?: string;
        id?: (number) | null;
        modified_at?: string;
        role?: string;
        username?: string;
    }>;
};

/**
 * Event schema
 */
export type Event = {
    JobMessages?: Array<{
        host?: string;
        msg?: string;
        redfish_error?: {
            code?: string;
            error?: {
                '@Message.ExtendedInfo'?: Array<{
                    Message?: string;
                    'MessageArgs.@odata.count'?: number;
                    MessageId?: string;
                    'RelatedProperties.@odata.count'?: number;
                    Resolution?: string;
                    Severity?: string;
                }>;
                code?: string;
                message?: string;
            };
        };
        status?: string;
    }>;
    Message?: string;
    Severity?: string;
    Time?: string;
    User?: string;
};

/**
 * GenericResponse schema
 */
export type GenericResponse = {
    changed?: number;
    detail?: string;
    title?: string;
};

/**
 * GetRolesResponse schema
 */
export type GetRolesResponse = {
    roles?: Array<{
        name?: string;
        permission_list?: Array<{
            method?: string;
            path?: string;
        }>;
        unassigned_permission_list?: Array<{
            method?: string;
            path?: string;
        }>;
    }>;
};

/**
 * HTTPError schema
 */
export type HTTPError = {
    /**
     * Human readable error message
     */
    detail?: (string) | null;
    errors?: Array<{
        more?: {
            [key: string]: unknown;
        };
        name?: string;
        reason?: string;
    }> | null;
    instance?: (string) | null;
    /**
     * HTTP status code
     */
    status?: (number) | null;
    /**
     * Short title of the error
     */
    title?: (string) | null;
    /**
     * URL of the error type. Can be used to lookup the error in a documentation
     */
    type?: (string) | null;
};

/**
 * Host schema
 */
export type Host = {
    bonds?: Array<({
        bmc?: boolean;
        fqdn?: string;
        id?: (number) | null;
        ifname?: string;
        ip?: string;
        mac?: string;
        mtu?: number;
        peers?: Array<(string)>;
        vlan?: string;
    } | null)>;
    boot_image?: string;
    firmware?: string;
    id?: (number) | null;
    interfaces?: Array<({
        bmc?: boolean;
        fqdn?: string;
        id?: (number) | null;
        ifname?: string;
        ip?: string;
        mac?: string;
        mtu?: number;
        vlan?: string;
    } | null)>;
    name?: string;
    provision?: boolean;
    tags?: Array<(string)>;
    uid?: (string) | null;
};

/**
 * JobMessage schema
 */
export type JobMessage = {
    host?: string;
    msg?: string;
    redfish_error?: {
        code?: string;
        error?: {
            '@Message.ExtendedInfo'?: Array<{
                Message?: string;
                'MessageArgs.@odata.count'?: number;
                MessageId?: string;
                'RelatedProperties.@odata.count'?: number;
                Resolution?: string;
                Severity?: string;
            }> | null;
            code?: string;
            message?: string;
        };
    };
    status?: string;
};

/**
 * NodeAddRequest schema
 */
export type NodeAddRequest = {
    node_list?: Array<({
        bonds?: Array<({
            bmc?: boolean;
            fqdn?: string;
            id?: (number) | null;
            ifname?: string;
            ip?: string;
            mac?: string;
            mtu?: number;
            peers?: Array<(string)>;
            vlan?: string;
        } | null)>;
        boot_image?: string;
        firmware?: string;
        id?: (number) | null;
        interfaces?: Array<({
            bmc?: boolean;
            fqdn?: string;
            id?: (number) | null;
            ifname?: string;
            ip?: string;
            mac?: string;
            mtu?: number;
            vlan?: string;
        } | null)>;
        name?: string;
        provision?: boolean;
        tags?: Array<(string)>;
        uid?: (string) | null;
    } | null)>;
};

/**
 * NodeBootImageRequest schema
 */
export type NodeBootImageRequest = {
    image?: string;
};

/**
 * NodeBootTokenResponse schema
 */
export type NodeBootTokenResponse = {
    nodes?: Array<{
        name?: string;
        token?: string;
    }>;
};

/**
 * NodeProvisionRequest schema
 */
export type NodeProvisionRequest = {
    provision?: boolean;
};

/**
 * NodeTagsRequest schema
 */
export type NodeTagsRequest = {
    /**
     * comma separated list of tags
     */
    tags?: string;
};

/**
 * PatchRolesRequest schema
 */
export type PatchRolesRequest = {
    permission_list?: Array<{
        method?: string;
        path?: string;
    }>;
    role?: string;
};

/**
 * PostRolesRequest schema
 */
export type PostRolesRequest = {
    inherited_role?: string;
    role?: string;
};

/**
 * RedfishJob schema
 */
export type RedfishJob = {
    jobs?: Array<({
        '@odata.context'?: string;
        '@odata.id'?: string;
        '@odata.type'?: string;
        CreatedBy?: string;
        Description?: string;
        EndTime?: string;
        EstimatedDuration?: string;
        HidePayload?: boolean;
        Id?: string;
        JobState?: string;
        JobStatus?: string;
        MaxExecutionTime?: string;
        Messages?: Array<{
            '@odata.id'?: string;
            Id?: string;
            Message?: string;
            MessageArgs?: Array<(string)>;
            MessageId?: string;
            Name?: string;
            RelatedProperties?: Array<((string) | null)> | null;
            Resolution?: string;
            Severity?: string;
        }>;
        Name?: string;
        Payload?: {
            HttpHeaders?: Array<((string) | null)> | null;
            HttpOperation?: string;
            JsonBody?: string;
            TargetUri?: string;
        };
        PercentComplete?: number;
        Schedule?: {
            EnabledDaysOfMonth?: Array<((number) | null)> | null;
            EnabledDaysOfWeek?: Array<((string) | null)> | null;
            EnabledIntervals?: Array<((string) | null)> | null;
            EnabledMonthsOfYear?: Array<((string) | null)> | null;
            InitialStartTime?: string;
            Lifetime?: string;
            MaxOccurrences?: number;
            RecurrenceInterval?: string;
        };
        StartTime?: string;
        StepOrder?: Array<((string) | null)> | null;
    } | null)>;
    name?: string;
};

/**
 * RedfishMetricReport schema
 */
export type RedfishMetricReport = {
    name?: string;
    reports?: Array<({
        '@odata.context'?: string;
        '@odata.etag'?: string;
        '@odata.id'?: string;
        '@odata.type'?: string;
        Context?: string;
        Description?: string;
        Id?: string;
        MetricValues?: Array<{
            MetricID?: string;
            MetricProperty?: string;
            MetricValue?: string;
            Oem?: unknown;
            Timestamp?: string;
        }>;
        Name?: string;
        Oem?: unknown;
        Timestamp?: string;
    } | null)>;
};

/**
 * RedfishSystem schema
 */
export type RedfishSystem = {
    bios_version?: string;
    boot_next?: string;
    boot_order?: Array<(string)>;
    health?: string;
    host_name?: string;
    manufacturer?: string;
    model?: string;
    name?: string;
    oem?: {
        Dell?: {
            DellSystem?: {
                ManagedSystemSize?: string;
                MaxCPUSockets?: number;
                MaxDIMMSlots?: number;
                MaxPCIeSlots?: number;
                SystemID?: number;
            };
        };
    };
    power_status?: string;
    processor_count?: number;
    serial_number?: string;
    total_memory?: number;
};

/**
 * User schema
 */
export type User = {
    created_at?: string;
    enabled?: boolean;
    hash?: string;
    id?: (number) | null;
    modified_at?: string;
    role?: string;
    username?: string;
};

/**
 * UserEnableRequest schema
 */
export type UserEnableRequest = {
    enabled?: boolean;
};

/**
 * UserRoleRequest schema
 */
export type UserRoleRequest = {
    /**
     * type of model.Role, valid options: disabled, user, admin
     */
    role?: string;
};

/**
 * UserStoreRequest schema
 */
export type UserStoreRequest = {
    password?: string;
    username?: string;
};

/**
 * UserStoreResponse schema
 */
export type UserStoreResponse = {
    role?: string;
    username?: string;
};

/**
 * unknown-interface schema
 */
export type unknown_interface = unknown;

export type PatchV1AuthResetData = {
    /**
     * Request body for api.AuthResetRequest
     */
    body: AuthResetRequest;
    headers?: {
        Accept?: string;
    };
};

export type PatchV1AuthResetResponse = (GenericResponse);

export type PatchV1AuthResetError = (HTTPError);

export type PostV1AuthSigninData = {
    /**
     * Request body for api.AuthRequest
     */
    body: AuthRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1AuthSigninResponse = (AuthResponse);

export type PostV1AuthSigninError = (HTTPError);

export type DeleteV1AuthSignoutData = {
    headers?: {
        Accept?: string;
    };
};

export type DeleteV1AuthSignoutResponse = (GenericResponse);

export type DeleteV1AuthSignoutError = (HTTPError);

export type PostV1AuthSignupData = {
    /**
     * Request body for api.AuthSignupRequest
     */
    body: AuthSignupRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1AuthSignupResponse = (AuthResponse);

export type PostV1AuthSignupError = (HTTPError);

export type PostV1AuthTokenData = {
    /**
     * Request body for api.AuthTokenRequest
     */
    body: AuthTokenRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1AuthTokenResponse = (AuthTokenReponse);

export type PostV1AuthTokenError = (HTTPError);

export type GetV1BmcData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type GetV1BmcResponse = (Array<RedfishSystem>);

export type GetV1BmcError = (HTTPError);

export type PostV1BmcConfigureAutoData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PostV1BmcConfigureAutoResponse = (Array<JobMessage>);

export type PostV1BmcConfigureAutoError = (HTTPError);

export type PostV1BmcConfigureImportData = {
    /**
     * Request body for api.BmcImportConfigurationRequest
     */
    body: BmcImportConfigurationRequest;
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PostV1BmcConfigureImportResponse = (Array<JobMessage>);

export type PostV1BmcConfigureImportError = (HTTPError);

export type GetV1BmcJobsData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type GetV1BmcJobsResponse = (Array<RedfishJob>);

export type GetV1BmcJobsError = (HTTPError);

export type DeleteV1BmcJobsJidsData = {
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * Redfish Job IDs. Use 'JID_CLEARALL' to clear all jobs
         */
        jids: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type DeleteV1BmcJobsJidsResponse = (Array<JobMessage>);

export type DeleteV1BmcJobsJidsError = (HTTPError);

export type GetV1BmcMetricsData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type GetV1BmcMetricsResponse = (Array<RedfishMetricReport>);

export type GetV1BmcMetricsError = (HTTPError);

export type PostV1BmcPowerBmcData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PostV1BmcPowerBmcResponse = (Array<JobMessage>);

export type PostV1BmcPowerBmcError = (HTTPError);

export type PostV1BmcPowerOsData = {
    /**
     * Request body for api.BmcOsPowerBody
     */
    body: BmcOsPowerBody;
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PostV1BmcPowerOsResponse = (Array<JobMessage>);

export type PostV1BmcPowerOsError = (HTTPError);

export type DeleteV1BmcSelData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type DeleteV1BmcSelResponse = (Array<JobMessage>);

export type DeleteV1BmcSelError = (HTTPError);

export type GetV1DbDumpData = {
    headers?: {
        Accept?: string;
    };
};

export type GetV1DbDumpResponse = (DataDump);

export type GetV1DbDumpError = (HTTPError);

export type PostV1DbRestoreData = {
    /**
     * Request body for model.DataDump
     */
    body: DataDump;
    headers?: {
        Accept?: string;
    };
};

export type PostV1DbRestoreResponse = (GenericResponse);

export type PostV1DbRestoreError = (HTTPError);

export type GetV1GrendelEventsData = {
    headers?: {
        Accept?: string;
    };
};

export type GetV1GrendelEventsResponse = (Array<Event>);

export type GetV1GrendelEventsError = (HTTPError);

export type DeleteV1ImagesData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by name
         */
        names?: string;
    };
};

export type DeleteV1ImagesResponse = (GenericResponse);

export type DeleteV1ImagesError = (HTTPError);

export type GetV1ImagesData = {
    headers?: {
        Accept?: string;
    };
};

export type GetV1ImagesResponse = (Array<BootImage>);

export type GetV1ImagesError = (HTTPError);

export type PostV1ImagesData = {
    /**
     * Request body for api.BootImageAddRequest
     */
    body: BootImageAddRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1ImagesResponse = (GenericResponse);

export type PostV1ImagesError = (HTTPError);

export type GetV1ImagesFindData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by name
         */
        names?: string;
    };
};

export type GetV1ImagesFindResponse = (Array<BootImage>);

export type GetV1ImagesFindError = (HTTPError);

export type DeleteV1NodesData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type DeleteV1NodesResponse = (GenericResponse);

export type DeleteV1NodesError = (HTTPError);

export type GetV1NodesData = {
    headers?: {
        Accept?: string;
    };
};

export type GetV1NodesResponse = (Array<Host>);

export type GetV1NodesError = (HTTPError);

export type PostV1NodesData = {
    /**
     * Request body for api.NodeAddRequest
     */
    body: NodeAddRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1NodesResponse = (GenericResponse);

export type PostV1NodesError = (HTTPError);

export type GetV1NodesFindData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type GetV1NodesFindResponse = (Array<Host>);

export type GetV1NodesFindError = (HTTPError);

export type PatchV1NodesImageData = {
    /**
     * Request body for api.NodeBootImageRequest
     */
    body: NodeBootImageRequest;
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PatchV1NodesImageResponse = (GenericResponse);

export type PatchV1NodesImageError = (HTTPError);

export type PatchV1NodesProvisionData = {
    /**
     * Request body for api.NodeProvisionRequest
     */
    body: NodeProvisionRequest;
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PatchV1NodesProvisionResponse = (GenericResponse);

export type PatchV1NodesProvisionError = (HTTPError);

export type PatchV1NodesTagsActionData = {
    /**
     * Request body for api.NodeTagsRequest
     */
    body: NodeTagsRequest;
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * option to add or remove tags
         */
        action: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type PatchV1NodesTagsActionResponse = (GenericResponse);

export type PatchV1NodesTagsActionError = (HTTPError);

export type GetV1NodesTokenInterfaceData = {
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * interface token will be created for
         */
        interface: string;
    };
    query?: {
        /**
         * Filter by nodeset. Minimum of one query parameter is required
         */
        nodeset?: string;
        /**
         * Filter by tags. Minimum of one query parameter is required
         */
        tags?: string;
    };
};

export type GetV1NodesTokenInterfaceResponse = (NodeBootTokenResponse);

export type GetV1NodesTokenInterfaceError = (HTTPError);

export type GetV1RolesData = {
    headers?: {
        Accept?: string;
    };
    query?: {
        /**
         * Filter by name
         */
        name?: string;
    };
};

export type GetV1RolesResponse = (GetRolesResponse);

export type GetV1RolesError = (HTTPError);

export type PatchV1RolesData = {
    /**
     * Request body for api.PatchRolesRequest
     */
    body: PatchRolesRequest;
    headers?: {
        Accept?: string;
    };
};

export type PatchV1RolesResponse = (GenericResponse);

export type PatchV1RolesError = (HTTPError);

export type PostV1RolesData = {
    /**
     * Request body for api.PostRolesRequest
     */
    body: PostRolesRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1RolesResponse = (GenericResponse);

export type PostV1RolesError = (HTTPError);

export type DeleteV1RolesNamesData = {
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * Delete by name
         */
        names: string;
    };
};

export type DeleteV1RolesNamesResponse = (GenericResponse);

export type DeleteV1RolesNamesError = (HTTPError);

export type GetV1UsersData = {
    headers?: {
        Accept?: string;
    };
};

export type GetV1UsersResponse = (Array<User>);

export type GetV1UsersError = (HTTPError);

export type PostV1UsersData = {
    /**
     * Request body for api.UserStoreRequest
     */
    body: UserStoreRequest;
    headers?: {
        Accept?: string;
    };
};

export type PostV1UsersResponse = (UserStoreResponse);

export type PostV1UsersError = (HTTPError);

export type DeleteV1UsersUsernamesData = {
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * target usernames
         */
        usernames: string;
    };
};

export type DeleteV1UsersUsernamesResponse = (GenericResponse);

export type DeleteV1UsersUsernamesError = (HTTPError);

export type PatchV1UsersUsernamesEnableData = {
    /**
     * Request body for api.UserEnableRequest
     */
    body: UserEnableRequest;
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * target usernames
         */
        usernames: string;
    };
};

export type PatchV1UsersUsernamesEnableResponse = (GenericResponse);

export type PatchV1UsersUsernamesEnableError = (HTTPError);

export type PatchV1UsersUsernamesRoleData = {
    /**
     * Request body for api.UserRoleRequest
     */
    body: UserRoleRequest;
    headers?: {
        Accept?: string;
    };
    path: {
        /**
         * target usernames
         */
        usernames: string;
    };
};

export type PatchV1UsersUsernamesRoleResponse = (GenericResponse);

export type PatchV1UsersUsernamesRoleError = (HTTPError);