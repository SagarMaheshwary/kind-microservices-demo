import { Controller, Logger } from '@nestjs/common';
import { Ctx, EventPattern, Payload, RmqContext } from '@nestjs/microservices';

@Controller()
export class NotificationController {
  private readonly logger = new Logger(NotificationController.name, {
    timestamp: true,
  });

  @EventPattern('user.created')
  async handleUserCreated(@Payload() data: any, @Ctx() context: RmqContext) {
    this.logger.log(`Sending welcome email to ${data.name}`);

    // Acknowledge message manually
    const channel = context.getChannelRef();
    const originalMessage = context.getMessage();
    channel.ack(originalMessage);
  }
}
