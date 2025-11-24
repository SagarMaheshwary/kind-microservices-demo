import { Injectable } from '@nestjs/common';
import { AmqpConnectionManager } from 'amqp-connection-manager';

export interface HealthStatus {
  status: string;
  details: {
    rabbitmq: string;
  };
}

@Injectable()
export class HealthService {
  private amqpManager: AmqpConnectionManager;

  setRMQ(manager: AmqpConnectionManager) {
    this.amqpManager = manager;
  }

  check() {
    const status: HealthStatus = {
      status: 'ready',
      details: { rabbitmq: 'ok' },
    };

    if (!this.amqpManager.isConnected()) {
      status.status = 'unready';
      status.details.rabbitmq = 'disconnected';
    }

    return status;
  }
}
