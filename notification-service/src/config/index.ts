export default () => ({
  server: {
    url: <string>getEnv('HTTP_SERVER_HOST', '0.0.0.0'),
    port: <number>getEnv('HTTP_SERVER_PORT', 4000),
  },
  amqp: {
    host: <string>getEnv('AMQP_HOST', 'rabbitmq'),
    port: <number>getEnv('AMQP_PORT', 5672),
    username: <string>getEnv('AMQP_USERNAME', 'default'),
    password: <string>getEnv('AMQP_PASSWORD', 'default'),
    queue: <string>getEnv('AMQP_QUEUE', 'notification-service'),
    maxConnectionAttempts: <number>getEnv('AMQP_MAX_CONNECTION_ATTEMPTS', 10),
  },
});

const getEnv = (key: string, defaultVal: any = null) => {
  return process.env[key] || defaultVal;
};
