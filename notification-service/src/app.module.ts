import { MiddlewareConsumer, Module } from '@nestjs/common';
import { HealthModule } from './modules/health/health.module';
import { LoggerMiddleware } from './middleware/logger.middleware';
import { NotificationModule } from './modules/notification/notification.module';
import { ConfigModule } from '@nestjs/config';
import config from './config';
import { existsSync } from 'fs';
import { join } from 'path';
import { GracefulShutdownModule } from 'nestjs-graceful-shutdown';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [config],
      ignoreEnvFile: !existsSync(join(__dirname, '..', '.env')),
    }),
    GracefulShutdownModule.forRoot({
      cleanup: async (app) => {
        for await (const microservice of app.getMicroservices()) {
          await microservice.close();
        }
        await app.close();
      },
    }),
    HealthModule,
    NotificationModule,
  ],
  controllers: [],
  providers: [],
})
export class AppModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(LoggerMiddleware).forRoutes('*');
  }
}
