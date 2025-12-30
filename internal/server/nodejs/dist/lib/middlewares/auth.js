"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const jsonwebtoken_1 = __importDefault(require("jsonwebtoken"));
const dotenv_1 = __importDefault(require("dotenv"));
dotenv_1.default.config();
function authMiddleware(req, res, next) {
    const token = req.headers['authorization'] || req.headers['Authorization'];
    if (!token) {
        res.status(401).json({ status: 'unauthorized', service: 'ai-chat' });
        return;
    }
    console.log(token);
    const tokenString = token.toString().split(' ')[1];
    console.log(tokenString);
    const decoded = jsonwebtoken_1.default.verify(tokenString, process.env.SESSION_KEY || '');
    if (!decoded) {
        res.status(401).json({ status: 'unauthorized', service: 'ai-chat' });
        return;
    }
    next();
}
exports.default = authMiddleware;
//# sourceMappingURL=auth.js.map