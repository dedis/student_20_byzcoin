#! /bin/python
import matplotlib.pyplot as plt
import pandas as pd
import os
import datetime

data_dir = './test_data/'
files = [(data_dir + fname) for fname in os.listdir(data_dir)\
         if fname.startswith('coins') and fname.endswith('.csv')]

def read_all_files(files):
    df = pd.DataFrame()
    for fname in files:
        data = pd.read_csv(fname)
        # We need to add variables regarding
        # batching and keeping here.
        batch = '_batch' in fname
        keep = not '_nokeep' in fname
        rowcount = len(data.index)
        b_vals = pd.Series([batch for i in range(rowcount)])
        k_vals = pd.Series([keep for i in range(rowcount)])
        data = data.assign(batch=b_vals.values)
        data = data.assign(keep=k_vals.values)
        # If the data frame is empty (first iteration),
        # we append no matter what. Otherwise, we append
        # IFF the colums are the same.
        if df.empty \
           or (len(data.columns) == len(df.columns) \
           and (data.columns == df.columns).all()):
            df = df.append(data, ignore_index=True)
    return df

df = read_all_files(files)
delays = list(set(df['delay']))
keep = list(set(df['keep']))
batch = list(set(df['batch']))

for delay in delays:
    for k in keep:
        for b in batch:
            titlestring =  'Byzcoin Simulation Monitor'
            # No whitespace, colons or commata in filenames
            namestring = (str(datetime.datetime.now()) + '-' + titlestring).replace(' ','').replace(':','-')

            data = df.loc[df['delay'] == delay].sort_values('hosts')
            data = data.loc[data['keep'] == k]
            data = data.loc[data['batch'] == b]
            data = data.reset_index()

            fig, axs = plt.subplots(1, 4)
            fig.set_size_inches(20, 5)

            fig.suptitle(titlestring)

            # [[ax.set_ylim([0, 70]) for ax in a] for a in axs]
            # [ax[0].set_ylabel('Time in seconds') for ax in axs]
            [ax.set_ylim([0, 55]) for ax in axs]
            axs[0].set_ylabel('Time in seconds')

            data.plot.bar(x='hosts', tick_label='overview', y= ['send_user_sum', 'prepare_user_sum', 'confirm_user_sum'], stacked=True, ax=axs[0])
            axs[0].set_xlabel('overview')
            plt.setp(axs[0].get_xticklabels(), visible=False)

            # data.plot.bar(x='hosts', y=['create_state_change_user_sum'], stacked=True, ax=axs[0][1])

            # data.plot.bar(x='hosts', y=['process_one_tx_user_sum'], stacked=True, ax=axs[0][2])

            data.plot.bar(x='hosts', y=['p_o_t.init_user_sum', 'p_o_t.execute_user_sum', 'p_o_t.increment_user_sum', 'p_o_t.verify_user_sum', 'p_o_t.store_user_sum'], stacked=True, ax=axs[1])
            axs[1].set_xlabel('ProcessOneTx')
            plt.setp(axs[1].get_xticklabels(), visible=False)

            # 'execute.Recover_user_sum' should also be in there
            data.plot.bar(x='hosts', y=['execute.newROSkipChain_user_sum', 'execute.GetValues_user_sum', 'execute.ContractConstructor_user_sum', 'execute.CreateContract_user_sum', 'execute.VerifyInstruction_user_sum', 'execute.Instruction_user_sum', 'execute.Trie_user_sum'], stacked=True, ax=axs[2])
            axs[2].set_xlabel('ExecuteInstruction')
            plt.setp(axs[2].get_xticklabels(), visible=False)

            # 'verify.ContractWrite_user_sum', 'verify.ContractCredential_user_sum', 'verify.ContractSpawner_user_sum', 'verify.contractAdaptorNV_user_sum', 'verify.contractNaming_user_sum', 'verify.ContractPopParty_user_sum', 'verify.ContractRoPaSci_user_sum', 'verify.contractAttrValue_user_sum', 'verify.contractDeferred_user_sum'
            # data.plot.bar(x='hosts', y=['verify.basicContract_user_sum', 'verify.contractConfig_user_sum'], stacked=True, ax=axs[1][2])

            data.plot.bar(x='hosts', y=['v_w_o.signers_user_sum', 'v_w_o.counters_user_sum', 'v_w_o.config_user_sum', 'v_w_o.darc_user_sum', 'v_w_o.action_user_sum', 'v_w_o.signatures_user_sum', 'v_w_o.check_user_sum', 'v_w_o.eval_user_sum'], stacked=True, ax=axs[3])
            axs[3].set_xlabel('VerifyWithOptions')
            plt.setp(axs[3].get_xticklabels(), visible=False)

            plt.savefig(data_dir + namestring + '.png')
            plt.close()
