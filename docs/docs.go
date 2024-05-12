package docs

const DocTemplate = `{
	"schemes": {{ marshal .Schemes }},
    "openapi": "3.0.0",
  	"info": {
		"description": "{{escape .Description}}",
    	"title": "{{.Title}}",
    	"contact": {},
        "version": "{{.Version}}"
  	},
	"host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
  	"paths": {
    "/ping": {
        "get": {
            "description": "Пинг БД",
            "tags": [
                "Ping"
            ],
            "summary": "Пинг",
            "parameters": [],
            "responses": {
                "204": {
                    "description": "Все в порядке"
                },
                "500": {
                    "description": "Внутренняя ошибка сервера",
                    "content": {
                      "application/json": {
                        "schema": {
                          "type": "object",
                          "properties": {
                            "error": {
                              "type": "string"
                            }
                          }
                        }
                      }
                    }
                }
            }
        }
    },
    "/commands": {
      "post": {
        "description": "Запрос для создания bash-команды (и запуска в горутине, для ограничения числа одновременно выполняющихся команд используется семафор, ограничение можно задать через .env файл)",
        "tags": [
          "Commands"
        ],
        "summary": "Запрос создания и запуска команды",
        "parameters": [],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "command": {
                    "type": "string",
                    "description": "Запускаемая команда",
                    "example": "ls"
                  },
                }
              }
            }
          }
        },
        "responses": {
          "202": {
            "description": "Создание прошло успешно, команда запустится когда позволит семафор",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "command_id": {
                        "type": "integer",
                        "description": "id команды",
                        "example": 1
                    }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      },
      "get": {
        "description": "Запрос для получения списка с информацией о командах. Можно использовать лимит и оффсет, по умолчанию используется лимит 15, оффест 0, а максимальный лимит - 50. Список отсортирован по id в порядке возрастания",
        "tags": [
            "Commands"
        ],
        "summary": "Запрос получения списка команд",
        "parameters": [
            {
                "in": "query",
                "name": "limit",
                "required": false,
                "schema": {
                  "type": "integer",
                  "description": "Лимит (максимальное количество элементов списка), максимум - 50, по умолчанию 15"
                }
            },
            {
                "in": "query",
                "name": "offset",
                "required": false,
                "schema": {
                  "type": "integer",
                  "description": "Оффсет (сдвиг), минимум - 0, по умолчанию - 0"
                }
            },
        ],
        "responses": {
          "200": {
            "description": "Успешно найдены элементы списка",
            "content": {
              "application/json": {
                "schema": {
                    "type": "array",
                    "items": {
                        "properties": {
                        "command_id": {
                            "type": "integer",
                            "description": "Идентификатор команды"
                        },
                        "command": {
                            "type": "string",
                            "description": "Команда"
                        },
                        "pid": {
                            "type": "integer",
                            "description": "PID процесса, выполняющего команду"
                        },
                        "output": {
                            "type": "string",
                            "description": "Вывод команды"
                        },
                        "status": {
                            "type": "string",
                            "description": "Статус выполнения команды"
                        },
                        "exitStatus": {
                            "type": "integer",
                            "description": "Exit код команды"
                        }
                        }
                    }
                  },
                "example": [
                    {
                        "command_id": 1,
                        "command": "ls",
                        "pid": 5,
                        "output": "abc",
                        "status": "done",
                        "exitStatus": 0
                    }
                ]
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "204": {
            "description": "Не найдены команды на \"странице\" с таким лимитом и оффсетом",
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/commands/stop/{command_id}": {
      "get": {
        "description": "Запрос для остановки выполенения команды, если ее можно остановить - то она найдется по PID в отдельной горутине и получит kill (используется singleflight, чтобу минимизировать число обращений к ОС)",
        "tags": [
            "Commands"
        ],
        "summary": "Запрос остановки выполнения команды",
        "parameters": [
            {
                "in": "path",
                "name": "command_id",
                "required": true,
                "schema": {
                  "type": "integer",
                  "description": "Идентификатор команды"
                }
            }
        ],
        "responses": {
          "202": {
            "description": "Запрос остановки зарегистрирован и будет обработан в отдельной горутине",
          },
          "400": {
            "description": "Некорректные данные/команда и так уже не запущена",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "404": {
            "description": "Команда с таким id не найдена",
            "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                        "error": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/commands/{command_id}": {
      "get": {
        "description": "Запрос для получения информации о команде по id",
        "tags": [
            "Commands"
        ],
        "summary": "Получение команды по id",
        "parameters": [
            {
                "in": "path",
                "name": "command_id",
                "required": true,
                "schema": {
                  "type": "integer",
                  "description": "Идентификатор команды"
                }
            }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "command_id": {
                        "type": "integer",
                        "description": "Идентификатор команды"
                    },
                    "command": {
                        "type": "string",
                        "description": "Команда"
                    },
                    "pid": {
                        "type": "integer",
                        "description": "PID процесса, выполняющего команду"
                    },
                    "output": {
                        "type": "string",
                        "description": "Вывод команды"
                    },
                    "status": {
                        "type": "string",
                        "description": "Статус выполнения команды"
                    },
                    "exitStatus": {
                        "type": "integer",
                        "description": "Exit код команды"
                    }
                  }
                },
                "example": {
                    "command_id": 1,
                    "command": "ls",
                    "pid": 5,
                    "output": "abc",
                    "status": "done",
                    "exitStatus": 0
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "404": {
            "description": "Команда с таким id не найдена",
            "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                        "error": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      },
    },
    "/commands/output/{command_id}": {
        "get": {
          "description": "Запрос для получения вывода команды по id",
          "tags": [
              "Commands"
          ],
          "summary": "Получение вывода команды по id",
          "parameters": [
              {
                  "in": "path",
                  "name": "command_id",
                  "required": true,
                  "schema": {
                    "type": "integer",
                    "description": "Идентификатор команды"
                  }
              }
          ],
          "responses": {
            "200": {
              "description": "OK",
                "content": {
                    "text/plain": {}
                }
            },
            "204": {
                "description": "Вывод команды пуст"
            },
            "400": {
              "description": "Некорректные данные",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "error": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            },
            "404": {
              "description": "Команда с таким id не найдена",
              "content": {
                  "application/json": {
                    "schema": {
                      "type": "object",
                      "properties": {
                          "error": {
                                  "type": "string"
                              }
                          }
                      }
                  }
              }
            },
            "500": {
              "description": "Внутренняя ошибка сервера",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "error": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        },
    },
  }
}`
