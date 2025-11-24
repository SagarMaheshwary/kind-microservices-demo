import { Injectable, Logger, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

@Injectable()
export class LoggerMiddleware implements NestMiddleware {
  private readonly logger = new Logger('HTTP', { timestamp: true });

  use(req: Request, res: Response, next: NextFunction) {
    const start = Date.now();

    const { method, originalUrl } = req;
    const query = req.query;
    const clientIp =
      req.headers['x-forwarded-for'] || req.socket.remoteAddress || req.ip;

    // Continue request and wait for response to finish
    res.on('finish', () => {
      const latency = Date.now() - start;
      const status = res.statusCode;

      const logMessage = {
        message: 'incoming request',
        method,
        path: originalUrl,
        query,
        client_ip: clientIp,
        status,
        latency: `${latency}ms`,
      };

      if (status >= 400) {
        this.logger.error('incoming request', JSON.stringify(logMessage));
      } else {
        this.logger.log(JSON.stringify(logMessage));
      }
    });

    next();
  }
}
