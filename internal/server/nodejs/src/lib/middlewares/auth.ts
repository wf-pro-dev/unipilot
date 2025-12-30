import { NextFunction,  Response, Request } from "express";
import jwt from 'jsonwebtoken';
import dotenv from 'dotenv';

dotenv.config();

function authMiddleware(req: Request, res: Response, next: NextFunction) {

    const token = req.headers['authorization'] || req.headers['Authorization'];
    if (!token) {
        res.status(401).json({ status: 'unauthorized', service: 'ai-chat' });
        return;
    }
    const tokenString: string = token.toString().split(' ')[1];
    const decoded = jwt.verify(tokenString, process.env.SESSION_KEY || '');
    if (!decoded) {
        res.status(401).json({ status: 'unauthorized', service: 'ai-chat' });
        return;
    }
    next();
}

export default authMiddleware;