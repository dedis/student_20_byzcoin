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

def plot_columns(ax, columns):
    last = None
    for column in columns:
        ax.bar(ind, data[column], label=column[:-9], bottom=last)
        last = data[column]

    ax.set_xlabel('transactions / batch size')
    ax.legend(bbox_to_anchor=(0.1, 0.5), loc='center left',)

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

            ind = ("2/5")

            fig, axs = plt.subplots(1, 2)

            fig.suptitle(titlestring)

            plot_columns(axs[0], ['prepare_user_sum',
                                  'send_user_sum',
                                  'confirm_user_sum'])

            axs[0].set_ylabel('Time in seconds')
            
            #data.plot(y='full_simulation_user_sum', marker='o')

            #data.plot.bar(
            #        y=[
            #            'prepare_user_sum',
            #            'send_user_sum',
            #            'confirm_user_sum',
            #            #'collect_tx_user_sum',
            #            #'process_tx_user_sum',
            #            #'create_new_block_user_sum',
            #            #'create_state_change_user_sum',
            #            #'update_trie_callback_user_sum',
            #            #'process_one_tx_user_sum'
            #        ], stacked=True)

            plot_columns(axs[1], ['create_new_block_user_sum',
                                  'create_state_change_user_sum',
                                  'update_trie_callback_user_sum',
                                  'process_one_tx_user_sum'])

            #data.plot(y='full_simulation_user_sum', marker='o')

            #data.plot.bar(
            #        y=[
            #            #'prepare_user_sum',
            #            #'send_user_sum',
            #            #'confirm_user_sum',
            #            #'collect_tx_user_sum',
            #            #'process_tx_user_sum',
            #            'create_new_block_user_sum',
            #            'create_state_change_user_sum',
            #            'update_trie_callback_user_sum',
            #            'process_one_tx_user_sum'
            #        ], stacked=True)

            plt.savefig(data_dir + namestring + '.png')
            plt.close()
