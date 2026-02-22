"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.QueryService = void 0;
const jmespath_1 = __importDefault(require("jmespath"));
class QueryService {
    apply(data, expression) {
        return jmespath_1.default.search(data, expression);
    }
}
exports.QueryService = QueryService;
