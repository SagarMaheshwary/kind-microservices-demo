import { NestFactory } from '@nestjs/core';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import { AppModule } from './app.module';
import { ConsoleLogger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { HealthService } from './modules/health/health.service';
import { setupGracefulShutdown } from 'nestjs-graceful-shutdown';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, {
    logger: new ConsoleLogger({
      colors: false,
      json: true,
    }),
  });

  setupGracefulShutdown({ app });

  const config = app.get<ConfigService>(ConfigService);

  const amqpConfig = config.get('amqp');
  const rmq = app.connectMicroservice<MicroserviceOptions>({
    transport: Transport.RMQ,
    options: {
      urls: [
        `amqp://${amqpConfig.username}:${amqpConfig.password}@${amqpConfig.host}:${amqpConfig.port}`,
      ],
      queue: amqpConfig.queue,
      queueOptions: {
        durable: false,
      },
      noAck: false,
      maxConnectionAttempts: amqpConfig.maxConnectionAttempts,
    },
  });

  await app.startAllMicroservices();

  const healthService = app.get<HealthService>(HealthService);
  healthService.setRMQ(
    rmq.unwrap<import('amqp-connection-manager').AmqpConnectionManager>(),
  );

  await app.listen(config.get('server.port'), config.get('server.host'));
}
bootstrap();
