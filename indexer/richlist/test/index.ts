import { Coin, LCDClient, Tx } from '@terra-money/terra.js';
import { APIRequester } from '@terra-money/terra.js/dist/client/lcd/APIRequester';

const beginHeight = 1;
const breakHeight = 0;
const interval = 1;
const richlistUrl = "http://localhost:1317";

const terra = new LCDClient({
    chainID: 'phoenix-1',
    URL: 'http://localhost:1317',
});

/**
 * A goal of this program is to validate balances of the addresses in the richlist at specified height, not to find a missing address.
 */
async function main() {
    let height = beginHeight;
    do {
        //console.log(`begin to validate ${height}`);
        const list = await getRichlist(height);
        const dup: string[] = [];

        for (const ranker of list.table) {
            const score = Coin.fromData(ranker.score);

            for (const address of ranker.address) {
            	//if (address=="terra190ggr8ahpa2kf7xd7mmngg4qlrs5m75570tcp4") continue;
                if (dup.find(elem => (elem == address))) {
                    throw new Error(`DUPLICATED ADDRESS FOUND! height:${height} address:${address}`)
                }
                dup.push(address);

                const [balances, _] = await terra.bank.balance(address, { height: height });
                const balance = balances.get('uluna');

                if (!(score.amount.equals(balance!.amount))) {
                    throw new Error(`INVALID SCORE DETECTED! height:${height} address:${address} balance:${balance?.toString()} / score:${score.toString()}`);
                } else {
                    ;//console.log(`${address} at ${height} is valid.`)
                }
            }
        }

        console.log(`[RESULT] height ${height} is valid.`);
        height += interval;
    } while (0 >= breakHeight || breakHeight >= height);
}

async function getRichlist(height?: number): Promise<Richlist> {
    const url = `/index/richlist/${height ?? "latest"}`;
    const requester = new APIRequester(richlistUrl)
    return await requester.get<Richlist>(url);
}

main().catch(console.error);

interface Ranker {
    position: number;
    address: string[];
    score: Coin.Data;
}
interface Richlist {
    height: number;
    table: Ranker[];
}
