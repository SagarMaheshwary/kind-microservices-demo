import {
  Controller,
  Get,
  HttpCode,
  HttpStatus,
  Inject,
  InternalServerErrorException,
  ServiceUnavailableException,
} from '@nestjs/common';
import { HealthService } from './health.service';

@Controller()
export class HealthController {
  constructor(
    @Inject(HealthService) private readonly healthService: HealthService,
  ) {}

  @Get('/livez')
  @HttpCode(HttpStatus.OK)
  async livez() {
    try {
      return { status: 'ok' };
    } catch (err) {
      throw new InternalServerErrorException(err);
    }
  }

  @Get('/readyz')
  @HttpCode(HttpStatus.OK)
  async readyz() {
    try {
      const status = this.healthService.check();
      if (status.status != 'ready') {
        throw new ServiceUnavailableException(status);
      }

      return status;
    } catch (err) {
      if (err instanceof ServiceUnavailableException) {
        throw err;
      }
      throw new InternalServerErrorException(err);
    }
  }
}
