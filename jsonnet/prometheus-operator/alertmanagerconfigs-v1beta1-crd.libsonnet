// Code generated using 'make generate-crds'. DO NOT EDIT.
{ spec+: { versions+: [
  {
    name: 'v1beta1',
    schema: {
      openAPIV3Schema: {
        properties: {
          apiVersion: {
            type: 'string',
          },
          kind: {
            type: 'string',
          },
          metadata: {
            type: 'object',
          },
          spec: {
            properties: {
              inhibitRules: {
                items: {
                  properties: {
                    equal: {
                      items: {
                        type: 'string',
                      },
                      type: 'array',
                    },
                    sourceMatch: {
                      items: {
                        properties: {
                          matchType: {
                            enum: [
                              '!=',
                              '=',
                              '=~',
                              '!~',
                            ],
                            type: 'string',
                          },
                          name: {
                            minLength: 1,
                            type: 'string',
                          },
                          value: {
                            type: 'string',
                          },
                        },
                        required: [
                          'name',
                        ],
                        type: 'object',
                      },
                      type: 'array',
                    },
                    targetMatch: {
                      items: {
                        properties: {
                          matchType: {
                            enum: [
                              '!=',
                              '=',
                              '=~',
                              '!~',
                            ],
                            type: 'string',
                          },
                          name: {
                            minLength: 1,
                            type: 'string',
                          },
                          value: {
                            type: 'string',
                          },
                        },
                        required: [
                          'name',
                        ],
                        type: 'object',
                      },
                      type: 'array',
                    },
                  },
                  type: 'object',
                },
                type: 'array',
              },
              receivers: {
                items: {
                  properties: {
                    discordConfigs: {
                      items: {
                        properties: {
                          apiURL: {
                            properties: {
                              key: {
                                type: 'string',
                              },
                              name: {
                                type: 'string',
                              },
                              optional: {
                                type: 'boolean',
                              },
                            },
                            required: [
                              'key',
                            ],
                            type: 'object',
                            'x-kubernetes-map-type': 'atomic',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          title: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    emailConfigs: {
                      items: {
                        properties: {
                          authIdentity: {
                            type: 'string',
                          },
                          authPassword: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          authSecret: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          authUsername: {
                            type: 'string',
                          },
                          from: {
                            type: 'string',
                          },
                          headers: {
                            items: {
                              properties: {
                                key: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                value: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'key',
                                'value',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          hello: {
                            type: 'string',
                          },
                          html: {
                            type: 'string',
                          },
                          requireTLS: {
                            type: 'boolean',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          smarthost: {
                            type: 'string',
                          },
                          text: {
                            type: 'string',
                          },
                          tlsConfig: {
                            properties: {
                              ca: {
                                properties: {
                                  configMap: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  secret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              cert: {
                                properties: {
                                  configMap: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  secret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              insecureSkipVerify: {
                                type: 'boolean',
                              },
                              keySecret: {
                                properties: {
                                  key: {
                                    type: 'string',
                                  },
                                  name: {
                                    type: 'string',
                                  },
                                  optional: {
                                    type: 'boolean',
                                  },
                                },
                                required: [
                                  'key',
                                ],
                                type: 'object',
                                'x-kubernetes-map-type': 'atomic',
                              },
                              serverName: {
                                type: 'string',
                              },
                            },
                            type: 'object',
                          },
                          to: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    name: {
                      minLength: 1,
                      type: 'string',
                    },
                    opsgenieConfigs: {
                      items: {
                        properties: {
                          actions: {
                            type: 'string',
                          },
                          apiKey: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          apiURL: {
                            type: 'string',
                          },
                          description: {
                            type: 'string',
                          },
                          details: {
                            items: {
                              properties: {
                                key: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                value: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'key',
                                'value',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          entity: {
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          note: {
                            type: 'string',
                          },
                          priority: {
                            type: 'string',
                          },
                          responders: {
                            items: {
                              properties: {
                                id: {
                                  type: 'string',
                                },
                                name: {
                                  type: 'string',
                                },
                                type: {
                                  enum: [
                                    'team',
                                    'teams',
                                    'user',
                                    'escalation',
                                    'schedule',
                                  ],
                                  minLength: 1,
                                  type: 'string',
                                },
                                username: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'type',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          source: {
                            type: 'string',
                          },
                          tags: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    pagerdutyConfigs: {
                      items: {
                        properties: {
                          class: {
                            type: 'string',
                          },
                          client: {
                            type: 'string',
                          },
                          clientURL: {
                            type: 'string',
                          },
                          component: {
                            type: 'string',
                          },
                          description: {
                            type: 'string',
                          },
                          details: {
                            items: {
                              properties: {
                                key: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                value: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'key',
                                'value',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          group: {
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          pagerDutyImageConfigs: {
                            items: {
                              properties: {
                                alt: {
                                  type: 'string',
                                },
                                href: {
                                  type: 'string',
                                },
                                src: {
                                  type: 'string',
                                },
                              },
                              type: 'object',
                            },
                            type: 'array',
                          },
                          pagerDutyLinkConfigs: {
                            items: {
                              properties: {
                                alt: {
                                  type: 'string',
                                },
                                href: {
                                  type: 'string',
                                },
                              },
                              type: 'object',
                            },
                            type: 'array',
                          },
                          routingKey: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          serviceKey: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          severity: {
                            type: 'string',
                          },
                          url: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    pushoverConfigs: {
                      items: {
                        properties: {
                          expire: {
                            pattern: '^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$',
                            type: 'string',
                          },
                          html: {
                            type: 'boolean',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          priority: {
                            type: 'string',
                          },
                          retry: {
                            pattern: '^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$',
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          sound: {
                            type: 'string',
                          },
                          title: {
                            type: 'string',
                          },
                          token: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          url: {
                            type: 'string',
                          },
                          urlTitle: {
                            type: 'string',
                          },
                          userKey: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    slackConfigs: {
                      items: {
                        properties: {
                          actions: {
                            items: {
                              properties: {
                                confirm: {
                                  properties: {
                                    dismissText: {
                                      type: 'string',
                                    },
                                    okText: {
                                      type: 'string',
                                    },
                                    text: {
                                      minLength: 1,
                                      type: 'string',
                                    },
                                    title: {
                                      type: 'string',
                                    },
                                  },
                                  required: [
                                    'text',
                                  ],
                                  type: 'object',
                                },
                                name: {
                                  type: 'string',
                                },
                                style: {
                                  type: 'string',
                                },
                                text: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                type: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                url: {
                                  type: 'string',
                                },
                                value: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'text',
                                'type',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          apiURL: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          callbackId: {
                            type: 'string',
                          },
                          channel: {
                            type: 'string',
                          },
                          color: {
                            type: 'string',
                          },
                          fallback: {
                            type: 'string',
                          },
                          fields: {
                            items: {
                              properties: {
                                short: {
                                  type: 'boolean',
                                },
                                title: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                value: {
                                  minLength: 1,
                                  type: 'string',
                                },
                              },
                              required: [
                                'title',
                                'value',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          footer: {
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          iconEmoji: {
                            type: 'string',
                          },
                          iconURL: {
                            type: 'string',
                          },
                          imageURL: {
                            type: 'string',
                          },
                          linkNames: {
                            type: 'boolean',
                          },
                          mrkdwnIn: {
                            items: {
                              type: 'string',
                            },
                            type: 'array',
                          },
                          pretext: {
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          shortFields: {
                            type: 'boolean',
                          },
                          text: {
                            type: 'string',
                          },
                          thumbURL: {
                            type: 'string',
                          },
                          title: {
                            type: 'string',
                          },
                          titleLink: {
                            type: 'string',
                          },
                          username: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    snsConfigs: {
                      items: {
                        properties: {
                          apiURL: {
                            type: 'string',
                          },
                          attributes: {
                            additionalProperties: {
                              type: 'string',
                            },
                            type: 'object',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          phoneNumber: {
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          sigv4: {
                            properties: {
                              accessKey: {
                                properties: {
                                  key: {
                                    type: 'string',
                                  },
                                  name: {
                                    type: 'string',
                                  },
                                  optional: {
                                    type: 'boolean',
                                  },
                                },
                                required: [
                                  'key',
                                ],
                                type: 'object',
                                'x-kubernetes-map-type': 'atomic',
                              },
                              profile: {
                                type: 'string',
                              },
                              region: {
                                type: 'string',
                              },
                              roleArn: {
                                type: 'string',
                              },
                              secretKey: {
                                properties: {
                                  key: {
                                    type: 'string',
                                  },
                                  name: {
                                    type: 'string',
                                  },
                                  optional: {
                                    type: 'boolean',
                                  },
                                },
                                required: [
                                  'key',
                                ],
                                type: 'object',
                                'x-kubernetes-map-type': 'atomic',
                              },
                            },
                            type: 'object',
                          },
                          subject: {
                            type: 'string',
                          },
                          targetARN: {
                            type: 'string',
                          },
                          topicARN: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    telegramConfigs: {
                      items: {
                        properties: {
                          apiURL: {
                            type: 'string',
                          },
                          botToken: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          chatID: {
                            format: 'int64',
                            type: 'integer',
                          },
                          disableNotifications: {
                            type: 'boolean',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          parseMode: {
                            enum: [
                              'MarkdownV2',
                              'Markdown',
                              'HTML',
                            ],
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    victoropsConfigs: {
                      items: {
                        properties: {
                          apiKey: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          apiUrl: {
                            type: 'string',
                          },
                          customFields: {
                            items: {
                              properties: {
                                key: {
                                  minLength: 1,
                                  type: 'string',
                                },
                                value: {
                                  type: 'string',
                                },
                              },
                              required: [
                                'key',
                                'value',
                              ],
                              type: 'object',
                            },
                            type: 'array',
                          },
                          entityDisplayName: {
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          messageType: {
                            type: 'string',
                          },
                          monitoringTool: {
                            type: 'string',
                          },
                          routingKey: {
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          stateMessage: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    webexConfigs: {
                      items: {
                        properties: {
                          apiURL: {
                            pattern: '^https?://.+$',
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          roomID: {
                            minLength: 1,
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                        },
                        required: [
                          'roomID',
                        ],
                        type: 'object',
                      },
                      type: 'array',
                    },
                    webhookConfigs: {
                      items: {
                        properties: {
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          maxAlerts: {
                            format: 'int32',
                            minimum: 0,
                            type: 'integer',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          url: {
                            type: 'string',
                          },
                          urlSecret: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                    wechatConfigs: {
                      items: {
                        properties: {
                          agentID: {
                            type: 'string',
                          },
                          apiSecret: {
                            properties: {
                              key: {
                                minLength: 1,
                                type: 'string',
                              },
                              name: {
                                minLength: 1,
                                type: 'string',
                              },
                            },
                            required: [
                              'key',
                              'name',
                            ],
                            type: 'object',
                          },
                          apiURL: {
                            type: 'string',
                          },
                          corpID: {
                            type: 'string',
                          },
                          httpConfig: {
                            properties: {
                              authorization: {
                                properties: {
                                  credentials: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  type: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                              basicAuth: {
                                properties: {
                                  password: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  username: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                },
                                type: 'object',
                              },
                              bearerTokenSecret: {
                                properties: {
                                  key: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                  name: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'key',
                                  'name',
                                ],
                                type: 'object',
                              },
                              followRedirects: {
                                type: 'boolean',
                              },
                              oauth2: {
                                properties: {
                                  clientId: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  clientSecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  endpointParams: {
                                    additionalProperties: {
                                      type: 'string',
                                    },
                                    type: 'object',
                                  },
                                  scopes: {
                                    items: {
                                      type: 'string',
                                    },
                                    type: 'array',
                                  },
                                  tokenUrl: {
                                    minLength: 1,
                                    type: 'string',
                                  },
                                },
                                required: [
                                  'clientId',
                                  'clientSecret',
                                  'tokenUrl',
                                ],
                                type: 'object',
                              },
                              proxyURL: {
                                type: 'string',
                              },
                              tlsConfig: {
                                properties: {
                                  ca: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  cert: {
                                    properties: {
                                      configMap: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                      secret: {
                                        properties: {
                                          key: {
                                            type: 'string',
                                          },
                                          name: {
                                            type: 'string',
                                          },
                                          optional: {
                                            type: 'boolean',
                                          },
                                        },
                                        required: [
                                          'key',
                                        ],
                                        type: 'object',
                                        'x-kubernetes-map-type': 'atomic',
                                      },
                                    },
                                    type: 'object',
                                  },
                                  insecureSkipVerify: {
                                    type: 'boolean',
                                  },
                                  keySecret: {
                                    properties: {
                                      key: {
                                        type: 'string',
                                      },
                                      name: {
                                        type: 'string',
                                      },
                                      optional: {
                                        type: 'boolean',
                                      },
                                    },
                                    required: [
                                      'key',
                                    ],
                                    type: 'object',
                                    'x-kubernetes-map-type': 'atomic',
                                  },
                                  serverName: {
                                    type: 'string',
                                  },
                                },
                                type: 'object',
                              },
                            },
                            type: 'object',
                          },
                          message: {
                            type: 'string',
                          },
                          messageType: {
                            type: 'string',
                          },
                          sendResolved: {
                            type: 'boolean',
                          },
                          toParty: {
                            type: 'string',
                          },
                          toTag: {
                            type: 'string',
                          },
                          toUser: {
                            type: 'string',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                  },
                  required: [
                    'name',
                  ],
                  type: 'object',
                },
                type: 'array',
              },
              route: {
                properties: {
                  activeTimeIntervals: {
                    items: {
                      type: 'string',
                    },
                    type: 'array',
                  },
                  continue: {
                    type: 'boolean',
                  },
                  groupBy: {
                    items: {
                      type: 'string',
                    },
                    type: 'array',
                  },
                  groupInterval: {
                    type: 'string',
                  },
                  groupWait: {
                    type: 'string',
                  },
                  matchers: {
                    items: {
                      properties: {
                        matchType: {
                          enum: [
                            '!=',
                            '=',
                            '=~',
                            '!~',
                          ],
                          type: 'string',
                        },
                        name: {
                          minLength: 1,
                          type: 'string',
                        },
                        value: {
                          type: 'string',
                        },
                      },
                      required: [
                        'name',
                      ],
                      type: 'object',
                    },
                    type: 'array',
                  },
                  muteTimeIntervals: {
                    items: {
                      type: 'string',
                    },
                    type: 'array',
                  },
                  receiver: {
                    type: 'string',
                  },
                  repeatInterval: {
                    type: 'string',
                  },
                  routes: {
                    items: {
                      'x-kubernetes-preserve-unknown-fields': true,
                    },
                    type: 'array',
                  },
                },
                type: 'object',
              },
              timeIntervals: {
                items: {
                  properties: {
                    name: {
                      type: 'string',
                    },
                    timeIntervals: {
                      items: {
                        properties: {
                          daysOfMonth: {
                            items: {
                              properties: {
                                end: {
                                  maximum: 31,
                                  minimum: -31,
                                  type: 'integer',
                                },
                                start: {
                                  maximum: 31,
                                  minimum: -31,
                                  type: 'integer',
                                },
                              },
                              type: 'object',
                            },
                            type: 'array',
                          },
                          months: {
                            items: {
                              pattern: '^((?i)january|february|march|april|may|june|july|august|september|october|november|december|[1-12])(?:((:((?i)january|february|march|april|may|june|july|august|september|october|november|december|[1-12]))$)|$)',
                              type: 'string',
                            },
                            type: 'array',
                          },
                          times: {
                            items: {
                              properties: {
                                endTime: {
                                  pattern: '^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)',
                                  type: 'string',
                                },
                                startTime: {
                                  pattern: '^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)',
                                  type: 'string',
                                },
                              },
                              type: 'object',
                            },
                            type: 'array',
                          },
                          weekdays: {
                            items: {
                              pattern: '^((?i)sun|mon|tues|wednes|thurs|fri|satur)day(?:((:(sun|mon|tues|wednes|thurs|fri|satur)day)$)|$)',
                              type: 'string',
                            },
                            type: 'array',
                          },
                          years: {
                            items: {
                              pattern: '^2\\d{3}(?::2\\d{3}|$)',
                              type: 'string',
                            },
                            type: 'array',
                          },
                        },
                        type: 'object',
                      },
                      type: 'array',
                    },
                  },
                  type: 'object',
                },
                type: 'array',
              },
            },
            type: 'object',
          },
        },
        required: [
          'spec',
        ],
        type: 'object',
      },
    },
    served: true,
    storage: false,
  },
] } }
