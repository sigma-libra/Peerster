//Author: Sabrina Kall
package gossiper

const FILE_FOLDER = "./_SharedFiles/"
const DOWNLOAD_FOLDER = "./_Downloads/"
const CHUNK_SIZE = 8 * 1024

const PACKET_SIZE = CHUNK_SIZE + 1024
const SHA_SIZE = 32

const HOP_LIMIT = 10


const STATUS_COUNTDOWN_TIME = 10
const DOWNLOAD_COUNTDOWN_TIME = 5